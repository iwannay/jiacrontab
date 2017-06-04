package store

import (
	"fmt"
	"testing"
)

func TestStore_Get(t *testing.T) {

	type args struct {
		key  string
		want interface{}
	}

	store := NewStore("../.data/store.json")
	chan1 := make(chan int, 100)
	for i := 0; i < 100; i++ {
		chan1 <- i
		t.Run(fmt.Sprintf("num_%d", i), func(t *testing.T) {
			store.Wrap(func(s *Store) {
				n := <-chan1
				s.Data[fmt.Sprintf("type_%d", n)] = map[string]interface{}{
					fmt.Sprintf("human_%d", n): map[string]interface{}{
						"gender": "boy",
						"name":   fmt.Sprintf("john_%d", n),
						"home": map[string]interface{}{
							"address": "shanghai",
							"member": map[string]interface{}{
								"wife": fmt.Sprintf("huahua_%d", n),
							},
						},
					},
				}
			})
		})
	}

	store.Sync()
	tests := []args{
		args{
			key:  "type_2.human_2.home.member.wife",
			want: "huahua_2",
		},
		args{
			key:  "type_4.human_4.home.member.wife",
			want: "huahua_4",
		},
		args{
			key:  "type_3.human_4.home.member.wife",
			want: nil,
		},
	}

	for _, vv := range tests {
		ret := store.Get(vv.key)
		if ret.value != vv.want {
			t.Fatalf("store.Get(%s) want %s but %s", vv.key, vv.want, ret.value)
		}

		ret2 := store.GetCopy(vv.key)
		if ret.value != vv.want {
			t.Fatalf("store.GetCopy(%s) want %s but %s", vv.key, vv.want, ret2.value)
		}
	}

}

func TestStore_Load(t *testing.T) {
	type args struct {
		key  string
		want interface{}
	}
	store := NewStore("../.data/store.json")

	ret := store.Load()
	if ret.err != nil {
		t.Fatal(ret.err)
	}
	tests := []args{
		args{
			key:  "type_2.human_2.home.member.wife",
			want: "huahua_2",
		},
		args{
			key:  "type_4.human_4.home.member.wife",
			want: "huahua_4",
		},
		args{
			key:  "type_3.human_4.home.member.wife",
			want: nil,
		},
	}
	for _, vv := range tests {
		ret := store.Get(vv.key)
		if ret.value != vv.want {
			t.Fatalf("store.Get(%s) want %s but %s", vv.key, vv.want, ret.value)
		}

		ret2 := store.GetCopy(vv.key)
		if ret.value != vv.want {
			t.Fatalf("store.GetCopy(%s) want %s but %s", vv.key, vv.want, ret2.value)
		}
	}
}
