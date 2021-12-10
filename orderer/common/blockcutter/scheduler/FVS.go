package scheduler

import (
	"log"
	"sort"
)

// var logger *logging.Logger

// func init() {
// 	logger = flogging.MustGetLogger(pkgLogID)
// }

type FVS interface {
	Run() (int32, []bool)
	Unblock(v int32, blocked []bool, blockedMap *[][]int32)
	FindCycles(component *SCC) ([][]int32, [][]int32, [][]int, int32)
	FindCyclesRecur(component *SCC, explore []bool, startV, currentV int32, blocked []bool, stack *[]int32, blockedMap *[][]int32, cycles *[][]int32, cyclesMap *[][]int32, sumArray *[][]int, sum *int32) bool
	BreakCycles(component *SCC) []bool
}

type fvs struct {
	sccs            []SCC
	graph           *[][]int32
	invalidVertices []bool
	nvertices       int
	removeSetSize   int32
	allNodes        *[]*Node
}

func NewFVS(graph *[][]int32, nodes *[]*Node) FVS {
	return &fvs{
		sccs:            nil,
		graph:           graph,
		invalidVertices: make([]bool, len(*graph)),
		nvertices:       len(*graph),
		removeSetSize:   0,
		allNodes:        nodes,
	}
}

func (f *fvs) Run() (int32, []bool) {
	invalidVertices := make([]bool, f.nvertices)

	sccGen := NewTarjanSCC(f.graph)
	sccGen.SCC()

	for _, scc := range sccGen.GetSCCs() {
		inv := f.BreakCycles(&scc)
		for _, vertex := range scc.Vertices {
			invalidVertices[vertex] = inv[vertex]
		}
	}
	for i := 0; i < f.nvertices; i++ {
		if invalidVertices[i] {
			f.removeSetSize += 1
		}
	}

	return f.removeSetSize, invalidVertices
}

// Run Johnson's algorithm to find all cycles within this strongly connected component
// Return a matrix
// |  |v1|v2| ... |vn|
// |c1| 0| 1| ... | 1|
// |c2| 1| 0| ... | 0|
// Emits one row per cycle indicating whether a given vertex vk is a part of this cycle
//
//
// Also returns a array containing a number (sk) corresponding to the number of cycles which
// contains the given vertex (vk)
// |  |v1|v2| ... |vn|
// |  |s1|s2| ... |sk|
func (f *fvs) FindCycles(component *SCC) ([][]int32, [][]int32, [][]int, int32) {
	if len((*component).Vertices) == 1 {
		return nil, nil, nil, int32(0)
	}

	explore := make([]bool, f.nvertices)
	copy(explore, (*component).Member)

	sum := int32(0)
	cycles := make([][]int32, 0, 1024)
	cyclesMap := make([][]int32, 0, 20)
	sumArray := make([][]int, f.nvertices)

	if len((*component).Vertices) == 2 {
		// SCC has only two vertices
		// Must contain a single cycle
		sum += 2
		cycle := make([]int32, 2)
		cycleBool := make([]int32, f.nvertices)
		cycle[0], cycle[1] = (*component).Vertices[0], (*component).Vertices[1]
		cycleBool[(*component).Vertices[0]], cycleBool[(*component).Vertices[1]] = 1, 1
		sumArray[(*component).Vertices[0]] = append(sumArray[(*component).Vertices[0]], len(cycles))
		sumArray[(*component).Vertices[1]] = append(sumArray[(*component).Vertices[1]], len(cycles))
		cycles = append(cycles, cycle)
		cyclesMap = append(cyclesMap, cycleBool)
	} else {
		for _, v := range (*component).Vertices {

			stack := make([]int32, 0, len((*component).Vertices))
			blocked := make([]bool, f.nvertices)
			blockedMap := make([][]int32, f.nvertices)

			for i := 0; i < f.nvertices; i++ {
				blockedMap[i] = make([]int32, 0, len((*component).Vertices))
			}

			f.FindCyclesRecur(component, explore, v, v, blocked, &stack, &blockedMap, &cycles, &cyclesMap, &sumArray, &sum)
			explore[v] = false
		}

	}

	return cycles, cyclesMap, sumArray, sum
}

func (f *fvs) FindCyclesRecur(component *SCC, explore []bool, startV, currentV int32, blocked []bool, stack *[]int32, blockedMap *[][]int32, cycles *[][]int32, cyclesMap *[][]int32, sumArray *[][]int, sum *int32) bool {
	foundCycle := false
	*stack = append(*stack, currentV)
	blocked[currentV] = true

	for _, n := range (*(f.graph))[currentV] {
		if explore[n] == false {
			continue
		} else if n == startV {
			// found a cycle
			foundCycle = true
			cycle := make([]int32, 0, len(*stack))
			cycleBool := make([]int32, f.nvertices)
			for _, iter := range *stack {
				(*sum) += 1
				cycleBool[iter] = 1
				(*sumArray)[iter] = append((*sumArray)[iter], len(*cycles))
				cycle = append(cycle, iter)
			}
			*cycles = append(*cycles, cycle)
			*cyclesMap = append(*cyclesMap, cycleBool)
		} else if blocked[n] == false {
			ret := f.FindCyclesRecur(component, explore, startV, n, blocked, stack, blockedMap, cycles, cyclesMap, sumArray, sum)
			foundCycle = foundCycle || ret
		}
	}

	if foundCycle {
		// recursive unblock currentV
		f.Unblock(currentV, blocked, blockedMap)

	} else {
		for _, v := range (*(f.graph))[currentV] {
			if explore[v] {
				(*blockedMap)[v] = append((*blockedMap)[v], currentV)
			}
		}
	}

	// stack pop()
	*stack = (*stack)[:len(*stack)-1]
	return foundCycle

}

func (f *fvs) Unblock(v int32, blocked []bool, blockedMap *[][]int32) {
	blocked[v] = false
	for i := 0; i < len((*blockedMap)[v]); i++ {
		n := (*blockedMap)[v][i]
		if blocked[n] {
			f.Unblock(n, blocked, blockedMap)
		}
	}
	(*blockedMap)[v] = nil
}

func (f *fvs) BreakCycles(component *SCC) []bool {
	invalidVertices := make([]bool, f.nvertices)
	circles, _, sumArray, _ := f.FindCycles(component)

	// fmt.Println(cycles)
	// phase 1
	n := len(circles)
	vis := make([]bool, n)
	for i := 0; i < n; i++ {
		if vis[i] == false {
			vis[i] = true
			min := int(1e9)
			idx := -1
			for j := 0; j < len(circles[i]); j++ {
				if min > (*f.allNodes)[circles[i][j]].weight {
					idx = int(circles[i][j])
					min = (*f.allNodes)[circles[i][j]].weight
				}
			}
			for j := 0; j < len(circles[i]); j++ {
				(*f.allNodes)[circles[i][j]].weight -= min
			}
			invalidVertices[idx] = true

			for _, cid := range sumArray[idx] {
				vis[cid] = true
				// for _, cur := range circles[cid] {
				// 		(*f.allNodes)[cur].weight -= min
				// 	}
			}
		}
	}

	// check circles
	existCircle := func() bool {
		var dfs func(cur int) bool
		vis := make([]int, f.nvertices)
		dfs = func(cur int) bool {
			vis[cur] = 1
			for _, v := range (*f.graph)[cur] {
				if vis[v] == 1 {
					return true
				}
				if component.Member[v] && vis[v] == 0 && invalidVertices[v] == false {
					if dfs(int(v)) == true {
						return true
					}
				}
			}
			vis[cur] = 2
			return false
		}
		for i := 0; i < len(component.Vertices); i++ {
			id := component.Vertices[i]
			if vis[id] == 0 && invalidVertices[id] == false {
				log.Printf("debug v2: dfs start from %d", id)
				if dfs(int(id)) == true {
					// found circle
					return true
				}
			}
		}
		return false
	}
	// debug
	if existCircle() {
		panic("debug v2 circle!!")
	}

	// sort deleted nodes by weight
	type weightedNode struct {
		weight int
		idx    int
	}
	var deleted []weightedNode
	for i, idx := range invalidVertices {
		if idx == true {
			deleted = append(deleted, weightedNode{
				weight: len((*f.allNodes)[i].txids),
				idx:    i,
			})
		}
	}
	sort.SliceStable(deleted, func(i, j int) bool {
		return deleted[i].weight > deleted[j].weight
	})
	// phase 2
	for _, node := range deleted {
		invalidVertices[node.idx] = false
		if existCircle() {
			invalidVertices[node.idx] = true
		}
	}

	return invalidVertices
}
