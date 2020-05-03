package lib

import (
	"fmt"
	"gopoc/utils"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Req *http.Request
	Poc *Poc
}

func checkVul(tasks []Task, ticker *time.Ticker) <-chan Task {
	var wg sync.WaitGroup
	results := make(chan Task)
	for _, task := range tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			<-ticker.C
			isVul, err := executePoc(task.Req, task.Poc)
			if err != nil {
				utils.Error(err)
				return
			}
			if isVul {
				results <- task
			}
		}(task)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

func BatchCheckSinglePoc(targets []string, pocName string, rate int) {
	if p, err := LoadSinglePoc(pocName); err == nil {
		rateLimit := time.Second / time.Duration(rate)
		ticker := time.NewTicker(rateLimit)
		defer ticker.Stop()
		var tasks []Task
		for _, target := range targets {
			req, _ := http.NewRequest("GET", target, nil)
			task := Task{
				Req: req,
				Poc: p,
			}
			tasks = append(tasks, task)
		}
		for result := range checkVul(tasks, ticker) {
			fmt.Println(result.Req.URL, result.Poc.Name)
		}
	}
}

func BatchCheckMultiPoc(targets []string, pocName string, threads, rate int) {
	pocs := LoadMultiPoc(pocName)
	rateLimit := time.Second / time.Duration(rate)
	ticker := time.NewTicker(rateLimit)
	defer ticker.Stop()

	in := make(chan string)
	go func() {
		for _, target := range targets {
			in <- target
		}
		close(in)
	}()

	worker := func(targets <-chan string, wg *sync.WaitGroup, retCh chan<- []Task) {
		defer wg.Done()
		for target := range targets {
			var tasks []Task
			var results []Task
			req, _ := http.NewRequest("GET", target, nil)
			for _, poc := range pocs {
				task := Task{
					Req: req,
					Poc: poc,
				}
				tasks = append(tasks, task)
			}
			for result := range checkVul(tasks, ticker) {
				results = append(results, result)
			}
			retCh <- results
		}
	}

	do := func() <-chan []Task {
		var wg sync.WaitGroup
		retCh := make(chan []Task, threads)
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go worker(in, &wg, retCh)
		}
		go func() {
			wg.Wait()
			close(retCh)
		}()
		return retCh
	}
	for results := range do() {
		for _, result := range results {
			utils.Green("%s %s", result.Req.URL, result.Poc.Name)
		}
	}
}

func CheckSinglePoc(req *http.Request, pocName string) *Poc {
	if p, err := LoadSinglePoc(pocName); err == nil {
		if isVul, err := executePoc(req, p); err == nil {
			if isVul {
				return p
			}
		}
	}
	return nil
}

func CheckMultiPoc(req *http.Request, pocName string, rate int) {
	rateLimit := time.Second / time.Duration(rate)
	ticker := time.NewTicker(rateLimit)
	defer ticker.Stop()
	var tasks []Task
	for _, poc := range LoadMultiPoc(pocName) {
		task := Task{
			Req: req,
			Poc: poc,
		}
		tasks = append(tasks, task)
	}
	for result := range checkVul(tasks, ticker) {
		utils.Green("%s %s", result.Req.URL, result.Poc.Name)
	}
}

func executePoc(oReq *http.Request, p *Poc) (bool, error) {
	utils.Debug(oReq.URL.String(), p.Name)
	c := NewEnvOption()
	c.UpdateCompileOptions(p.Set)
	env, err := NewEnv(&c)
	if err != nil {
		utils.ErrorF("environment creation error: %s\n", err)
		return false, err
	}
	variableMap := make(map[string]interface{})
	req, err := ParseRequest(oReq)
	if err != nil {
		utils.Error(err)
		return false, err
	}
	variableMap["request"] = req

	// 现在假定set中payload作为最后产出，那么先排序解析其他的自定义变量，更新map[string]interface{}后再来解析payload
	newSet := make(map[string]string)
	keys := make([]string, 0)
	for k := range p.Set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		newSet[k] = p.Set[k]
	}
	for k, expression := range newSet {
		if k != "payload" {
			out, err := Calculate(env, expression, variableMap)
			if err != nil {
				continue
			}
			switch value := out.Value().(type) {
			case *UrlType:
				variableMap[k] = ParseUrlType(value)
			case int64:
				variableMap[k] = int(value)
			default:
				variableMap[k] = fmt.Sprintf("%v", out)
			}
		}
	}

	if p.Set["payload"] != "" {
		out, err := Calculate(env, p.Set["payload"], variableMap)
		if err != nil {
			return false, err
		}
		variableMap["payload"] = fmt.Sprintf("%v", out)
	}

	success := false
	for _, rule := range p.Rules {
		for k1, v1 := range variableMap {
			_, isMap := v1.(map[string]string)
			if isMap {
				continue
			}
			value := fmt.Sprintf("%v", v1)
			for k2, v2 := range rule.Headers {
				rule.Headers[k2] = strings.ReplaceAll(v2, "{{"+k1+"}}", value)
			}
			rule.Path = strings.ReplaceAll(strings.TrimSpace(rule.Path), "{{"+k1+"}}", value)
			rule.Body = strings.ReplaceAll(strings.TrimSpace(rule.Body), "{{"+k1+"}}", value)
		}

		if oReq.URL.Path != "" && oReq.URL.Path != "/" {
			req.Url.Path = fmt.Sprint(oReq.URL.Path, rule.Path)
		} else {
			req.Url.Path = rule.Path
		}
		newRequest, _ := http.NewRequest(rule.Method, fmt.Sprintf("%s://%s%s", req.Url.Scheme, req.Url.Host, req.Url.Path), strings.NewReader(rule.Body))
		newRequest.Header = oReq.Header.Clone()
		for k, v := range rule.Headers {
			newRequest.Header.Set(k, v)
		}
		resp, err := DoRequest(newRequest, rule.FollowRedirects)
		if err != nil {
			return false, err
		}
		//parseResponse(resp, variableMap)
		variableMap["response"] = resp

		if rule.Search != "" {
			result := doSearch(strings.TrimSpace(rule.Search), string(resp.Body))
			if result != nil && len(result) > 0 { // 正则匹配成功
				for k, v := range result {
					//fmt.Println(k, v)
					variableMap[k] = v
				}
				//return false, nil
			} else {
				return false, nil
			}
		}

		out, err := Calculate(env, rule.Expression, variableMap)
		if err != nil {
			return false, err
		}
		//fmt.Println(fmt.Sprintf("%v, %s", out, out.Type().TypeName()))
		if fmt.Sprintf("%v", out) == "false" { //如果false不继续执行后续rule
			success = false // 如果最后一步执行失败，就算前面成功了最终依旧是失败
			break
		}
		success = true
	}
	return success, nil
}

func doSearch(re string, body string) map[string]string {
	r, err := regexp.Compile(re)
	if err != nil {
		return nil
	}
	result := r.FindStringSubmatch(body)
	names := r.SubexpNames()
	if len(result) > 1 && len(names) > 1 {
		paramsMap := make(map[string]string)
		for i, name := range names {
			if i > 0 && i <= len(result) {
				paramsMap[name] = result[i]
			}
		}
		return paramsMap
	}
	return nil
}
