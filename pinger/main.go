package main

import (
        "flag"
        "fmt"
        "os"
        "os/exec"
        "os/signal"
        "strings"
        "sync"
        "time"

        "github.com/cheggaaa/pb/v3"
        "github.com/go-ping/ping"
)

//https://wiki.dieg.info/ulimit

func pingo(
        host string,
        timeout time.Duration,
        interval time.Duration,
        count, size, ttl int,
        privileged bool) int {

        pinger, err := ping.NewPinger(host)
        if err != nil {
                fmt.Println("ERROR:", err)
                return 500
        }
        // listen for ctrl-C signal
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)
        go func() {
                for range c {
                        pinger.Stop()
                }
        }()

        pinger.Count = count
        pinger.Size = size
        pinger.Interval = interval
        pinger.Timeout = timeout
        pinger.TTL = ttl
        pinger.SetPrivileged(privileged)

        err = pinger.Run()
        if err != nil {
                fmt.Println("Failed to ping target host:", err)
        }
        stats := pinger.Statistics()

        switch {
        case stats.PacketsSent == count && stats.PacketsRecv == 0:
                return 404 // не пингуется
        case stats.PacketsSent == count && stats.PacketsRecv < count:
                return 206 // Зафиксированы потери
        case stats.PacketsSent == count && stats.PacketsRecv == count:
                return 200 // Нормально пингуется
        }
        return 500 // Какая-то ошибка
}

func pingu(arg ...string) int {

        cmd := "fping"
        out, err := exec.Command(cmd, arg...).Output()
        if err != nil {
                //fmt.Println(string("ERR: "), err, arg)
                //log.Fatal(err)
        }
        contain := strings.Contains(string(out), "alive")
        if contain {
                //fmt.Print(string(out))
                return 200
        } else {
                return 404
        }
}

func masterPing(net string, threads int, timeout time.Duration, interval time.Duration, count int, size int, ttl int, privileged bool) {
        var count_goroutina int = 0
        start := time.Now()
        var rez int
        var ch = make(chan string, 10)
        var wg sync.WaitGroup
        // Запускаем несколько потоков...
        wg.Add(threads)
        n := 0

        for i := 0; i < threads; i++ {

                go func() {
                        for {
                                host, ok := <-ch
                                if !ok { // Когда задачи кончились и канал закрыт закрываем горутину
                                        wg.Done()
                                        //fmt.Println("Stop goroutina", a)
                                        count_goroutina++
                                        return
                                }
                                rez = pingo(host, timeout, interval, count, size, ttl, privileged) // Выполнение функции ping в отдельном потоке
                                //rez = pingu(host) // с использованием внешнего приложения fping
                                switch {
                                case rez == 500:
                                        //fmt.Println(host, "Внутренняя ошибка")
                                case rez == 200:
                                        n = n + 1
                                        //fmt.Println(host, "alive")
                                case rez == 404:
                                        //fmt.Println(host, "404")
                                case rez == 206:
                                        //fmt.Println(host, "Потери")
                                        n = n + 1
                                        //fmt.Println(host, "alive")
                                }
                        }
                }()
        }
        //fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

        // Now the jobs can be added to the channel, which is used as a queue
        f := generator_ip(net)
        //fmt.Printf("%v \n", *f.Next())
        //n := 0
        // 16 - 65534
        // 15 - 131070
        // 8 - 16777216
        bar := pb.StartNew(65534)
        for r := range f {
                ch <- r
                bar.Increment()
                //s := <-rezult
                //n = n + s
                //fmt.Printf("%v %v\n", n, r)
        }
        bar.Finish()
        close(ch) // Это говорит о том, что горутинам больше нечего делать
        wg.Wait() // Ждём завершения потоков
        //fmt.Println("всего горутин", count_goroutina)
        fmt.Println("Обнаружено хостов: ", n)
        duration := time.Since(start)
        fmt.Println("Процесс пингования занял:", duration)
}

var usage = `
Usage:

    ping [-c count] [-i interval] [-t timeout] [--privileged] host

Examples:

    # ping network continuously
    ping 10.228.0.0/16

    # ping google 5 times
    ping -c 5 10.228.0.0/16

    # ping network 5 times at 500ms intervals
    ping -c 5 -i 500ms 10.228.0.0/16

    # ping network for 10 seconds
    ping -t 10s www.google.com

    # Send a privileged raw ICMP ping
    sudo ping --privileged 212.220.0.0/24

    # Send ICMP messages with a 100-byte payload
    ping -s 100 212.220.0.0/24
`

func main() {

        //net := "212.220.0.0/24" // Сеть, которую надо пропинговать
        //var timeout time.Duration = time.Millisecond * 1000 * 2
        //var interval time.Duration = time.Millisecond * 500
        //var count int = 3 // количество пингов
        //var size int = 24 // размер пакета
        //var ttl int = 64  // время жизни
        //var privileged bool = false
        //var threads int = 1500 // Общее количество используемых потоков, за исключением основного main() потока

        threads := flag.Int("t", 100, "") // Общее количество используемых потоков, за исключением основного main() потока
        timeout := flag.Duration("o", time.Millisecond*1000*2, "")
        interval := flag.Duration("i", time.Millisecond*500, "")
        count := flag.Int("c", 3, "")
        size := flag.Int("s", 24, "")
        ttl := flag.Int("l", 64, "TTL")
        privileged := flag.Bool("privileged", false, "")
        flag.Usage = func() {
                fmt.Print(usage)
        }
        flag.Parse()

        if flag.NArg() == 0 {
                flag.Usage()
                return
        }

        network := flag.Arg(0)

        masterPing(network, *threads, *timeout, *interval, *count, *size, *ttl, *privileged)

        //bar := pb.StartNew(65534)
        //f := generator_ip("212.220.0.0/16")
        //for r := range f {
        //      bar.Increment()
        //      fmt.Println(r)
        //      time.Sleep(1000000)
        //}

}
