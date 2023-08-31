package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
)

func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	r, err := client.Put(context.Background(), "k1", "v1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", r.PrevKv)

	v, err := client.Get(context.Background(), "k1")
	if err != nil {
		panic(err)
	}
	fmt.Println("v1:", v.Kvs)
}
