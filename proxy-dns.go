package main

import (
  "bufio"
  "fmt"
  "os"
  "strings"
  "os/signal"
  "strings"
  "sync"
  "syscall"
  "github.com/miekg/dns"
)

///receives a filename containing a list of records to redirect to
func parse(filename string)(map[string]string,error){
  records := make(map[string]string)
  fh,err := os.Open(filename)
  if err != nil {
    return records, err
  }
  defer fh.Close()
  scanner := bufio.NewScanner(fh)
  for scanner.Scan(){
    line := scanner.Text()
    parts := strings.SplitN(line,",",2)
    if len(parts)<2{
      return records,fmt.Errorf("%s is not a valid line",line)
    }
    records[parts[0]] = parts[1]
  }
  return records,scanner.Err()
}

func main(){
  //sync it to auto read any changes to config file
  var recordLock sync.RWMutex
  records,err := parse("proxy.config")
  if err != nil {
    panic(err)
  }
  fmt.Printf("%+v\n",records)

  //create handler func
  dns.HandleFunc(".",func(w dns.ResponseWriter,req  *dns.Msg){
    if len(req.Question) < 1 {
      dns.HandleFailed(w,req)
      return
    }
    fqdn := req.Question[0].Name
    parts := strings.Split(fqdn,".")
    if len(parts) >= 2 {
      fqdn = strings.Join(psrts[len(parts)-2:],".")
    }
    recordLock.RLock()
    match,ok ;= records[fqdn]
    recordLock.RUnlock()
    if match == ""{
      dns.HandleFailed(w,req)
      return
    }
    resp,err := dns.Exchange(req,match)
    if err != nil {
      dns.HandleFailed(w,req)
      return
    }
/*    if != ok {
      dns.HandleFailed(w,req)
      return
    }
    if err := w.WriteMsg(resp);err != nil {
      dns.HandleFailed(w,req)
      return
    } */
  })
  go func(){
    sigs := make(chan os.Signal,1)
    signal.Notify(sigs,syscall.SIGUSR1)
    for sig := range sigs {
      switch sig {
      case syscall.SIGUSR1:
        log.Println("SIGUSR1: reloading records")
        recordLock.Lock()
        parse("proxy.config")
        recordLock.Unlock()
      }
    }
  }()
  log.Fatal(dns.ListenAndServe(":53","udp",nil))
}
