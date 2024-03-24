package homework_16

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

const WeightRoundRobin = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(WeightRoundRobin,
		&WeightedPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type WeightedPicker struct {
	mutex sync.Mutex
	conns []*weightConn
}

func (b *WeightedPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 这里实时计算 totalWeight 是为了方便你作业动态调整权重
	var totalWeight int
	var res *weightConn
	var resIndex int

	b.mutex.Lock()
	for i, node := range b.conns {
		totalWeight += node.weight
		node.currentWeight += node.weight
		if res == nil || res.currentWeight < node.currentWeight {
			res = node
			resIndex = i
		}
	}
	res.currentWeight -= totalWeight
	b.mutex.Unlock()
	return balancer.PickResult{
		SubConn: res.SubConn,
		Done: func(info balancer.DoneInfo) {
			if info.Err != nil {
				// 如果有错误，就降低节点权重
				st, ok := status.FromError(info.Err)
				if ok {
					// 如果是限流错误，就降低节点权重
					if st.Code() == codes.ResourceExhausted {
						res.weight -= 1
						if res.weight <= res.minWeight {
							res.weight = res.minWeight
						}
					} else if st.Code() == codes.Unavailable {
						// 如果是熔断错误，直接将节点挪出节点列表
						b.mutex.Lock()
						b.conns = append(b.conns[:resIndex], b.conns[resIndex+1:]...)
						b.mutex.Unlock()

						// 启动 goroutine 做健康检查
						go func() {
							b.healthCheck(res)
						}()
					}
				}
			} else {
				// 如果成功，就提高节点权重
				res.weight += 1
				if res.weight >= res.maxWeight {
					res.weight = res.maxWeight
				}
			}
		},
	}, nil
}

func (b *WeightedPicker) healthCheck(res *weightConn) {
	for {
		// 如果通过健康检查，就重新加入节点列表
		if res.healthCheck() {
			b.mutex.Lock()
			// 计算用的初始权重调整为初始权重的 1/10，避免流量过大
			res.weight = res.baseWeight/10 + 1
			b.conns = append(b.conns, res)
			b.mutex.Unlock()
		} else {
			// 间隔 1m 再做一次健康检查
			time.Sleep(time.Minute)
		}
	}
}

type WeightedPickerBuilder struct {
}

func (b *WeightedPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for con, conInfo := range info.ReadySCs {
		// 如果不存在，那么权重就是 0
		weightVal, _ := conInfo.Address.Metadata.(map[string]any)["weight"]
		// 经过注册中心的转发之后，变成了 float64，要小心这个问题
		weight, _ := weightVal.(float64)
		conns = append(conns, &weightConn{
			SubConn:       con,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}
	return &WeightedPicker{
		conns: conns,
	}
}

type weightConn struct {
	// 初始权重，不会改变
	baseWeight int
	// 初始权重
	weight int
	// 当前权重
	currentWeight int
	balancer.SubConn
	// 最大权重
	maxWeight int
	// 最小权重
	minWeight int
}

func (w *weightConn) healthCheck() bool {
	return true
}
