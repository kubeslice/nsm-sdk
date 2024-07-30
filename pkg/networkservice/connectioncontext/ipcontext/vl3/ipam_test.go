package vl3

import (
	"context"
	"fmt"
	"testing"
)

func TestAlloc(t *testing.T) {
	var vl3Ipam vl3IPAM
	vl3Ipam.reset(context.Background(), "10.6.16.0/20", []string{"192.168.0.0/16", "10.1.0.0/16"})
	ipNet, err := vl3Ipam.allocate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ipNet.String() != "10.6.16.1/32" {
		t.Errorf("Incorrect IP address allocated. Expected: %v, Got: %v", "10.6.16.1/32", ipNet.String())
	}
}

func TestAllocIPString(t *testing.T) {
	var vl3Ipam vl3IPAM
	vl3Ipam.reset(context.Background(), "10.6.16.0/20", []string{"192.168.0.0/16", "10.1.0.0/16"})
	ipNet, err := vl3Ipam.allocateIPString("10.6.16.7/32")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ipNet.String() != "10.6.16.7/32" {
		t.Errorf("Incorrect IP address allocated. Expected: %v, Got: %v", "10.6.16.7/32", ipNet.String())
	}
	ipNet, err = vl3Ipam.allocateIPString("10.6.16.7/32")
	if err == nil {
		t.Errorf("Allocating previously allocated IP")
	}
}

func TestAllocAndFreeIPSet(t *testing.T) {
	var vl3Ipam vl3IPAM
	vl3Ipam.reset(context.Background(), "10.6.16.0/20", []string{"192.168.0.0/16", "10.1.0.0/16"})
	for i := 0; i < 10; i++ {
	        ipNet, err := vl3Ipam.allocate()
	        if err != nil {
		        t.Errorf("Unexpected error: %v", err)
	        }
		fmt.Printf("Allocated IP: %s\n", ipNet.String())
	}
	for i := 0; i < 10; i++ {
		ip := fmt.Sprintf("10.6.16.%d", i + 1)
		_, err := vl3Ipam.allocateIPString(ip)
		if err == nil {
			t.Errorf("Previously allocated IP %v still available", ip)
		}
	}

	vl3Ipam.freeIPListAllocated([]string{"10.6.16.3/32", "10.6.16.5/32", "10.6.16.6/32"})
	if vl3Ipam.isExcluded("10.6.16.1/32") == false {
		t.Errorf("Freed wrong IP")
	}
	for _, ip := range []string{"10.6.16.3/32", "10.6.16.5/32", "10.6.16.6/32"} {
                if vl3Ipam.isExcluded(ip) == true {
			t.Errorf("Failed to free IP: %s", ip)
	        }
	}
}

func TestFreeIfAllocated(t *testing.T) {
	var vl3Ipam vl3IPAM
	vl3Ipam.reset(context.Background(), "10.6.16.0/20", []string{"192.168.0.0/16", "10.1.0.0/16"})
	ipNet, err := vl3Ipam.allocateIPString("10.6.16.7/32")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ipNet.String() != "10.6.16.7/32" {
		t.Errorf("Incorrect IP address allocated. Expected: %v, Got: %v", "10.6.16.7/32", ipNet.String())
	}

	vl3Ipam.freeIfAllocated(ipNet.String())

	if vl3Ipam.isExcluded("10.6.16.7/32") {
		t.Errorf("Failed to free IP address")
	}
}


