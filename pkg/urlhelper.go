package pkg

import (
	"strconv"
	"strings"
)

//IsURL 是否url
//	true -如:potato.im, mail@163.com
//	false -如:potato.xxx (.xxx不是域名后缀)
func IsURL(s string) bool {
	if len(s) < 4 { //域名最小长度为4
		return false
	}

	start := 0
	end := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c == '/' || c == '?' || c == '#' {
			end--
		} else {
			break
		}
	}

	//是否需要完全匹配域名后缀(非ip时)
	//当以http://或https://开头不需要匹配域名后缀
	//否则需要匹配域名后缀
	needForce := true

	//匹配http头
	if end > 7 && strings.EqualFold(s[:7], "http://") {
		needForce = false
		start = 7
	} else if end > 8 && strings.EqualFold(s[:8], "https://") {
		needForce = false
		start = 8
	} else if end > 5 && strings.EqualFold(s[:5], "wx://") {
		needForce = false
		start = 5
	} else if end > 4 && strings.EqualFold(s[:4], "git@") {
		return IsGitURL(s)
	}

	if end-start < 4 { //域名最小长度为4
		return false
	}

	var n = -1

	//匹配'/'或'?'
	for i := start; i < end; i++ {
		c := s[i]
		if c == '/' || c == '?' || c == '#' {
			n = i
			break
		}
	}

	//匹配host i要么大于3(域名必须大于3个字符),要么等于-1
	var host string

	if n == -1 {
		host = s[start:end]
	} else if n > 3 {
		host = s[start:n]
	} else {
		return false
	}

	//校验host及port
	var x []string //根据'.'拆分的切片
	i := strings.LastIndex(host, ":")
	if i > -1 {
		//port长度1-6位
		if len(host) <= i+1 || len(host) > i+6 {
			return false
		}
		//取port
		port := host[i+1:]
		//是否数字
		if !IsNumber(port) {
			return false
		}
		//转数字
		n, _ := strconv.Atoi(port)
		if n < 1 || n > 65535 {
			return false
		}
		x = strings.Split(host[:i], ".")
	} else {
		x = strings.Split(host, ".")
	}

	//拆分域名或ip
	if len(x) < 2 {
		return false
	}

	//取x最后一部分
	z := x[len(x)-1]

	//校验如果是ip地址
	if IsNumber(z) { //如果ip

		//ip必须是4部分,不支持ipv6
		if len(x) != 4 {
			return false
		}

		//判断ip里面前3部分,ip地址0-255
		for i := 0; i < 3; i++ {
			v := x[i]
			if len(v) <= 0 || len(v) > 3 || !IsNumber(v) {
				return false
			}
			n, e := strconv.Atoi(v)
			if e != nil {
				return false
			}
			if n < 0 || n > 255 {
				return false
			}
		}
	} else { //否则域名
		z = strings.ToLower(z) //转小写

		rLen := len(x)

		//如果强制,需要校验域名后缀
		if needForce {
			//校验域名后缀
			if !IsDomainSuffix("." + z) {
				return false
			}
			rLen--
		}

		//判断域名里面每一部分
		for i := 0; i < rLen; i++ {
			v := x[i]
			if len(v) == 0 {
				return false
			}

			//第一个字符或最后一个字符不能是连接符'-'
			if v[:1] == "-" || v[len(v)-1:] == "-" {
				return false
			}

			//处理邮箱格式
			if i == len(x)-2 {
				//倒数第二部分其中可以带@,如:邮箱格式 'mail@163.com'
				for i, c := range v {
					//第一个字符或最后一个字符不能是'@'
					if (i == 0 || i == len(v)-1) && c == '@' {
						return false
					}
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '-') || (c == '@')) {
						return false
					}
				}
			} else {
				//否则,组成必须是:26个英文字母、数字（0-9）以及连接符'-'
				for _, c := range v {
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '-')) {
						return false
					}
				}
			}
		}
	}
	return true
}

//IsGitUrl 是否git url如,git@gitlab.potato.im:gitlab.chatserver.im/interfaceprobuf.git
func IsGitURL(s string) bool {
	if len(s) < 4 {
		return false
	}
	if strings.EqualFold(s[:4], "git@") && strings.EqualFold(s[len(s)-4:], ".git") {
		n := len(s) - 4
		if i := strings.LastIndex(s, "/"); i > -1 {
			n = i
		}
		x := strings.Split(s[4:n], ":")
		for _, v := range x {
			y := strings.Split(v, ".")
			for _, v2 := range y {
				for i := 0; i < len(v2); i++ {
					c := v2[i]
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
						return false
					}
				}
			}
		}
		return true
	}
	return false
}

//IsAtUsername 是否@用户名(第一个字符必须是'@')
//	普通用户名:5-32个字符组成,必须是0-9,a-z,A-Z及下划线,且不能以下划线开始或结束
//	特殊用户名:'@all','@gif','@vid','@pic','@vote','@like','@imdb','@wiki','@bold'始终返回true
//	如:'@Zero7'返回true, '@Zero'返回false
func IsAtUsername(s string) bool {
	return len(s) > 1 && s[:1] == "@" && IsUsername(s[1:])
}

//IsUsername 是否用户名(不带@)用户名满足:[0-9a-zA-z_]且长度小于32
func IsUsername(s string) bool {
	switch {
	case len(s) == 3:
		return strings.EqualFold(s, "all") ||
			strings.EqualFold(s, "gif") ||
			strings.EqualFold(s, "vid") ||
			strings.EqualFold(s, "pic")
	case len(s) == 4:
		return strings.EqualFold(s, "vote") ||
			strings.EqualFold(s, "like") ||
			strings.EqualFold(s, "imdb") ||
			strings.EqualFold(s, "wiki") ||
			strings.EqualFold(s, "bold")
	case len(s) >= 5 && len(s) <= 32:
		for i := range s {
			c := s[i]

			//'_'不能是第一个或最后一个字符
			if c == '_' && (i == 0 || i == len(s)-1) {
				return false
			}

			//支持数字,英文字母及下划线
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '_')) {
				return false
			}
		}
		return true
	}
	return false
}

//isNumber 字符串是否全为数字
func IsNumber(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

//toNumber 字符串转数字
func ToNumber(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
