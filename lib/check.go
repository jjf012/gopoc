package lib

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func Check(oReq *http.Request, p *Poc) (bool, error) {
	c := NewEnvOption()
	c.UpdateCompileOptions(p.Set)
	env, err := NewEnv(&c)
	if err != nil {
		fmt.Printf("environment creation error: %s\n", err)
		return false, err
	}
	variableMap := make(map[string]interface{})
	req, err := ParseRequest(oReq)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	variableMap["request"] = req
	//variableMap := parseRequest(req)

	// 现在假定set中payload是最后产出，那么先排序解析其他的自定义变量，更新map[string]interface{}后再来解析payload
	newSet := make(map[string]string)
	keys := make([]string, 0)
	for k := range p.Set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		newSet[k] = p.Set[k]
	}
	for k, v := range newSet {
		if k != "payload" {
			out, err := Calculate(env, v, variableMap)
			if err != nil {
				continue
			}
			if u, ok := out.Value().(*UrlType); ok {
				variableMap[k] = ParseUrlType(u)
				continue
			}
			variableMap[k] = fmt.Sprintf("%v", out)
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

		req, _ := http.NewRequest(rule.Method, fmt.Sprintf("%s://%s%s", req.Url.Scheme, req.Url.Host, rule.Path), strings.NewReader(rule.Body))
		for k, v := range rule.Headers {
			req.Header.Set(k, v)
		}
		resp, err := DoRequest(req, rule.FollowRedirects)
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
