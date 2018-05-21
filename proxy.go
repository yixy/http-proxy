package main

import(
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"strconv"
	"strings"
	"fmt"
)

func proxyHandler(w http.ResponseWriter, r *http.Request){

	var requestHeader string
	for key,values := range r.Header{
		var v string
		for _,value:= range values{
			v+=value+";"
		}
		requestHeader += key + "[" + v[:len(v)-1]+"],"
	}

	proxyUrl:=r.Header.Get("Proxy-Url")

	if proxyUrl==""{
		fmt.Println("Proxy-Url can not be empty.")
		w.WriteHeader(511)
		_,err:=w.Write([]byte("Proxy-Url can not be empty."))
		if err !=nil{
			panic("proxyUrl can not be empty.")
		}
		return
	}

	rBody,err:=ioutil.ReadAll(r.Body)
	if err!=nil{
		fmt.Println("get request body error.",err)
		w.WriteHeader(513)
		_,err:=w.Write([]byte("get request body error."+err.Error()))
		if err !=nil{
			panic("get request body error.")
		}
		return
	}

	req,_:=http.NewRequest(r.Method, proxyUrl, strings.NewReader(string(rBody)))
	for k,v:=range r.Header{
		if k!= "Proxy-Url"{
			for _,vv:=range v{
				req.Header.Add(k,vv)
			}
		}
	}

	timeout,err:=strconv.Atoi(r.Header.Get("timeout-Set"))
	if(err!=nil || timeout<=0){
		timeout=20000
	}

	client:=&http.Client{
		Timeout:time.Duration(timeout)*time.Millisecond,
	}
	resp,err:=client.Do(req)

	if err!=nil{
		fmt.Println("call service exception.",err)
		w.WriteHeader(512)
		_,error:=w.Write([]byte("call service excption:"+err.Error()))
		if error !=nil{
			panic("call service exception!!!")
		}
		return
	}

	defer func(){
		resp.Body.Close()
	}()

	for k,v:=range resp.Header{
		for _,vv:=range v{
			w.Header().Add(k,vv)
		}
	}

	for _,value:=range resp.Request.Cookies(){
		w.Header().Add(value.Name,value.Value)
	}

	w.WriteHeader(resp.StatusCode)

	result,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		w.WriteHeader(514)
		_,err:=w.Write([]byte("read body err."+err.Error()))
		if err !=nil{
			panic("read body err.")
		}
		return
	}

	_,err=w.Write(result)
	if err!=nil{
		w.WriteHeader(515)
		_,err:=w.Write([]byte("write body err."+err.Error()))
		if err !=nil{
			panic("write body err.")
		}
		return
	}

	fmt.Println("-----------")
}

func main(){

	argNum:=len(os.Args)
	fmt.Println("Command params:")
	for i:=1;i<argNum;i++{
		fmt.Println(os.Args[i])
	}

	var port ="80"

	if argNum>1{
		port = os.Args[1]
	}

	http.HandleFunc("/proxy.do",proxyHandler)
	fmt.Println("Start serving on port:"+port)
	err:=http.ListenAndServe(":"+port,nil)
	if err!=nil{
		fmt.Println("start serving exception:",err)
	}
}

