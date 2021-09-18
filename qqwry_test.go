package qqwry

import (
	"log"
	"testing"
	"time"
)

func TestQQwry(t *testing.T) {
	q := NewQQwry("qqwry.dat")
	for i := 0; i < 30; i++ {
		go func() {
			q.Find("8.8.8.8")
			log.Printf("ip:%v, country:%v, city:%v", q.Ip, q.Country, q.City)
		}()
		go func() {
			q.Find("114.114.114.114")
			log.Printf("ip:%v, country:%v, city:%v", q.Ip, q.Country, q.City)
		}()
	}
	time.Sleep(20 * time.Second)
}
