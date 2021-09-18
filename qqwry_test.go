package qqwry

import (
	"log"
	"testing"
)

func TestQQwry(t *testing.T) {
	q := NewQQwry("qqwry.dat")
	q.Find("8.8.8.8")
	q.Find("114.114.114.114")
	log.Printf("ip:%v, country:%v, city:%v", q.Ip, q.Country, q.City)
}
