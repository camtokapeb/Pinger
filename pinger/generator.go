package main

import (
        "math"
        "net"
        "strconv"
        "strings"
)

type ipadress chan string

func inc(ip net.IP) {
        for j := len(ip) - 1; j >= 0; j-- {
                ip[j]++
                if ip[j] > 0 {
                        break
                }
        }
}

func (ip ipadress) Next() *string {
        c, ok := <-ip
        if !ok {
                return nil
        }
        return &c
}

func generator_ip(network string) ipadress {
        ip, ipnet, _ := net.ParseCIDR(network)
        mask := strings.Split(string(network), "/")
        zeros, _ := strconv.Atoi(mask[1])                       // Количество нулей в маске
        limit := math.Pow(2, float64(32-zeros)) - 1 // Количество хостов
        c := make(chan string)
        go func() {
                for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
                        if limit == 0 {
                                close(c)
                                return
                        }
                        a := ip.String()
                        c <- a
                        limit--
                }
        }()
        return c
}

//func main() {
//      f := generator_ip("192.168.0.0/25")
//      fmt.Printf("%v \n", *f.Next())
//      for r := range f {
//              fmt.Printf("%v \n", r)
//      }
//}
