//+build : trie

package dao

import (
	"fmt"
	"taobaoke/internal/model"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var trie = NewTrie()

func TestTrie_Insert(t *testing.T) {
	Convey("向trie中添加节点", t, func() {
		trie.Insert([]HandlerFunc{func(context *Context) {
			fmt.Println(1)
		}}, model.OrderFinish)
		So(trie.Size(), ShouldEqual, 1)
	})
}

func TestTrie_Insert2(t *testing.T) {
	Convey("递归向trie中添加节点", t, func() {
		trie.Insert2([]HandlerFunc{func(context *Context) {
			fmt.Println(1)
		}}, model.OrderFinish)
		So(trie.Size(), ShouldEqual, 1)
	})
}
func TestTrie_Search(t *testing.T) {
	Convey("查询这条状态链是否有对应处理方法", t, func() {
		trie.Insert([]HandlerFunc{func(context *Context) {
			fmt.Println(1)
		}}, model.OrderPaid, model.OrderFinish)
		contains := trie.Search(model.OrderPaid, model.OrderFinish)
		So(contains, ShouldBeTrue)
		contains = trie.Search(model.OrderPaid, model.OrderBalance)
		So(contains, ShouldBeFalse)
	})
}

func TestTrie_Search2(t *testing.T) {
	Convey("递归查询这条状态链是否有对应处理方法", t, func() {
		trie.Insert2([]HandlerFunc{func(context *Context) {
			fmt.Println(1)
		}}, model.OrderPaid, model.OrderFinish)
		contains := trie.Search2(model.OrderPaid, model.OrderFinish)
		So(contains, ShouldBeTrue)
		contains = trie.Search2(model.OrderPaid, model.OrderBalance)
		So(contains, ShouldBeFalse)
	})
}
func TestTrie_StartsWith(t *testing.T) {
	Convey("查询前缀", t, func() {
		trie.Insert([]HandlerFunc{func(context *Context) {
			fmt.Println(1)
		}}, model.OrderPaid, model.OrderFinish)
		ok := trie.StartsWith(model.OrderPaid)
		So(ok, ShouldBeTrue)
	})
}

func TestTrie_HandlerChain(t *testing.T) {
	Convey("获取责任链", t, func() {
		trie.Insert([]HandlerFunc{func(context *Context) {
			t.Log("UpdateOrderPaidStatus\n")
		}}, model.OrderCreate, model.OrderPaid)
		trie.Insert([]HandlerFunc{func(context *Context) {
			t.Log("UpdateOrderFinishStatus\n")
		}}, model.OrderCreate, model.OrderPaid, model.OrderFinish)
		trie.Insert([]HandlerFunc{func(context *Context) {
			t.Log("UpdateOrderBalanceStatus\n")
		}}, model.OrderCreate, model.OrderPaid, model.OrderBalance)
		trie.Insert([]HandlerFunc{func(context *Context) {
			t.Log("UpdateOrderBalanceStatus\n")
		}}, model.OrderCreate, model.OrderPaid, model.OrderFinish, model.OrderBalance)
		// paid --> balance
		chain := trie.HandlerChain(model.OrderCreate, model.OrderPaid)
		for _, fun := range chain {
			if fun != nil {
				fun(&Context{})
			}
		}
		// paid --> finish
		chain = trie.HandlerChain(model.OrderCreate, model.OrderPaid, model.OrderFinish)
		for _, fun := range chain {
			if fun != nil {
				fun(&Context{})
			}
		}
		// 0 --> balance
		chain = trie.HandlerChain(model.OrderCreate)
		for _, fun := range chain {
			if fun != nil {
				fun(&Context{})
			}
		}

	})
}
func TestAddHandle(t *testing.T) {
	a := func(ctx *Context) {
		fmt.Println("a")
	}
	b := func(ctx *Context) {
		fmt.Println("b")
	}
	c := func(ctx *Context) {
		fmt.Println("c")
	}
	d := func(ctx *Context) {
		fmt.Println("d")
	}
	m := map[model.OrderStatus]HandlerFunc{
		model.OrderCreate:  a,
		model.OrderPaid:    b,
		model.OrderFinish:  c,
		model.OrderBalance: d,
	}

	handles := []HandlerFunc{a, b, c, d}
	statuses := []model.OrderStatus{model.OrderCreate, model.OrderPaid, model.OrderFinish, model.OrderBalance}
	if len(statuses) != len(handles) {
		t.Fatal("状态和处理方法要一一对应")
	}
	//fn := func(f ...HandlerFunc) {
	//	for _, do := range f {
	//		do()
	//	}
	//}
	//handlesIndexSlices := subsets([]int{int(model.OrderCreate), int(model.OrderPaid), int(model.OrderFinish), int(model.OrderBalance)})
	statusesIndexSlices := subsets2(statuses)

	for ii, indexSlice := range statusesIndexSlices {
		h := make([]HandlerFunc, len(indexSlice))
		for i, v := range indexSlice {
			h[i] = m[v]
		}
		trie.Insert(h, statusesIndexSlices[ii]...)
	}
	// 输入原状态组,目标状态
	// 返回一条责任链

	chain := trie.HandlerChain(model.OrderCreate, model.OrderFinish)
	for _, fun := range chain {
		if fun != nil {
			fun(&Context{})
		}
	}
}

func subsets(nums []int) [][]int {
	ans := make([][]int, 0)

	var backtrace func(pos, num int, cur []int)
	backtrace = func(pos, num int, cur []int) {
		if len(cur) == num {
			tmp := make([]int, len(cur))
			copy(tmp, cur)
			ans = append(ans, tmp)
			return
		}

		for i := pos; i < len(nums); i++ {
			cur = append(cur, i)
			backtrace(i+1, num, cur)
			cur = cur[:len(cur)-1]
		}
	}

	for i := 0; i <= len(nums); i++ {
		cur := make([]int, 0, i)
		backtrace(0, i, cur)
	}

	return ans
}
func subsets2(nums []model.OrderStatus) [][]model.OrderStatus {
	ans := make([][]model.OrderStatus, 0)

	var backtrace func(pos, num int, cur []model.OrderStatus)
	backtrace = func(pos, num int, cur []model.OrderStatus) {
		if len(cur) == num {
			tmp := make([]model.OrderStatus, len(cur))
			copy(tmp, cur)
			ans = append(ans, tmp)
			return
		}

		for i := pos; i < len(nums); i++ {
			cur = append(cur, nums[i])
			backtrace(i+1, num, cur)
			cur = cur[:len(cur)-1]
		}
	}

	for i := 0; i <= len(nums); i++ {
		cur := make([]model.OrderStatus, 0, i)
		backtrace(0, i, cur)
	}

	return ans
}
