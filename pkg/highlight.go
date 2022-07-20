package pkg

import (
	"gitlab.moments.im/pkg/protoc/moment"
	"strings"
	"unicode/utf16"
)

type HighLightType int32

const (
	HighLightType_USERNAME HighLightType = iota + 1
	HighLightType_URL
)

//HighLight 高亮参数
//
//Offset : 字节偏移量
//Limit  : 字节长度
//UOffset: utf8偏移量
//ULimit : utf8长度
type HighLight struct {
	Type       HighLightType
	Offset     int32
	UserID     int32
	UserName   string
	AccessHash uint64
	Limit      int32
	UOffset    int32
	ULimit     int32
}

func GetAtHightLight(s string, entitys []*moment.HighLight) (atEntitys []*moment.HighLight) {
	atEntitys = make([]*moment.HighLight, 0, len(entitys))
	for _, entity := range entitys {
		if entity, s = GetAtHighLight(s, entity); entity == nil {
			continue
		}

		if len(atEntitys) > 0 {
			entity.Offset += atEntitys[len(atEntitys)-1].Offset + atEntitys[len(atEntitys)-1].Limit
			entity.UOffset += atEntitys[len(atEntitys)-1].UOffset + atEntitys[len(atEntitys)-1].ULimit
		}
		atEntitys = append(atEntitys, entity)
	}
	return
}

//GetAtHighLight 获取@ 高亮
func GetAtHighLight(s string, hl *moment.HighLight) (*moment.HighLight, string) {

	// special case
	if len(hl.UserName) == 0 {
		return nil, ""
	}
	var index int32
	var encodeContent []uint16
	if index = int32(strings.Index(s, hl.UserName)); index > -1 {

		hl.Type = moment.HighLightType_HighLightType_USERNAME

		hl.Offset = int32(index)

		hl.Limit = int32(len(hl.UserName))
		// fmt.Println(s[:hl.Offset])
		encodeContent = utf16.Encode([]rune(s[:hl.Offset]))

		hl.UOffset = int32(len(encodeContent))
		encodeContent = utf16.Encode([]rune(s[hl.Offset : hl.Offset+hl.Limit]))

		hl.ULimit = int32(len(encodeContent))
		s = s[index+int32(len(hl.UserName)):]
		return hl, s
	}

	return nil, s
}

//GetHighLights 高亮匹配
func GetHighLights(s string) []*HighLight {
	s = strings.Replace(s, "\n", " ", -1)
	rows := strings.Split(s, " ")

	hs := make([]*HighLight, 0)
	var gIndex int32
	var uIndex int32
	for _, row := range rows {
		//小于4直接跳过
		if len(row) < 4 {
			gIndex += int32(len(row)) + 1
			uIndex += int32(len(utf16.Encode([]rune(row)))) + 1
			continue
		}

		//每行开始匹配
		rs := regexpRow(row)
		for _, v := range rs {
			//计算Unicode offset,limit
			v.UOffset = int32(len(utf16.Encode([]rune(row[:v.Offset]))))
			v.ULimit = int32(len(utf16.Encode([]rune((row[v.Offset : v.Offset+v.Limit])))))

			v.Offset += gIndex
			v.UOffset += uIndex
		}
		hs = append(hs, rs...)
		gIndex += int32(len(row)) + 1
		uIndex += int32(len(utf16.Encode([]rune(row)))) + 1
	}
	return hs
}

//regexpRow
func regexpRow(row string) []*HighLight {
	if len(row) < 4 {
		return nil
	}

	hs := make([]*HighLight, 0)
	var regIndex int

	//nextRegFlag 1-匹配url;2-匹配username
	nextRegFlag := int8(1)

	//未匹配上url
	missRegUrl := false

	//未匹配上username
	// missRegUsername := false

	// if row[:1] == "@" {
	// 	nextRegFlag = 2
	// }

	//if row[:4] == "http" {
	//	nextRegFlag = 1
	//}

	//如果匹配到末尾或url,username都未匹配成功,就跳出循环
	for regIndex < len(row) && !(missRegUrl) {
		rs := row[regIndex:]

		switch nextRegFlag {
		case 1: //匹配url
			begin, end, ok := MatchURL(rs)
			if ok {
				//fmt.Println("url:", regIndex+begin, end-begin)
				hs = append(hs, &HighLight{
					Type:   HighLightType_URL,
					Offset: int32(regIndex + begin),
					Limit:  int32(end - begin),
				})
				regIndex += end
				missRegUrl = false
			} else {
				missRegUrl = true
			}
			nextRegFlag = 2 //下次匹配username

		// case 2: //匹配username
		// 	begin, end, ok := MatchUsername(rs)
		// 	if ok {
		// 		//fmt.Println("username:", regIndex+begin, end-begin)
		// 		hs = append(hs, &HighLight{
		// 			Type:   HighLightType_USERNAME,
		// 			Offset: int32(regIndex + begin),
		// 			Limit:  int32(end - begin),
		// 		})
		// 		regIndex += end
		// 		nextRegFlag = 2 //下次匹配username
		// 		missRegUsername = false
		// 	} else if begin > 0 {
		// 		nextRegFlag = 1 //下次匹配url
		// 		missRegUsername = true
		// 		missRegUrl = false
		// 		regIndex += begin + 1 //+1(这里包括@符号)
		// 	} else {
		// 		missRegUsername = true
		// 		nextRegFlag = 1 //下次匹配url
		// 	}
		default:
			//不会进到这里来
			return hs
		}
	}
	return hs
}

//MatchURL 如果匹配上,ok==true,并且返回起始位置
func MatchURL(s string) (start, end int, ok bool) {
	end = len(s) //default end is len(s)

	dotIndex := -1
	//找到第一个'.'
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			dotIndex = i
			break
		}
	}

	//如果未找到'.'或'.'为第一个字符,返回false
	if dotIndex <= 0 {
		return 0, 0, false
	}

	//然后以'.'向前匹配
	for i := dotIndex; i >= 0; i-- {
		c := s[i]
		//遇到其它字符跳出
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '-') || (c == ':') || (c == '/') || (c == '@') || (c == '.')) {
			start = i + 1 //不包含匹配到的字符
			break
		}
	}

	if end > start && IsURL(s[start:end]) {
		return start, end, true
	}
	return start, end, false
}

//MatchUsername 如果匹配上,ok==true,并且返回起始位置
func MatchUsername(s string) (start, end int, ok bool) {
	end = len(s) //default end is len(s)

	//匹配标识 -1:未匹配上, 1:匹配上@
	var matchFlag int8 = -1

	//匹配以@开始,以特殊字符结束
	for i := 0; i < len(s); i++ {
		c := s[i]
		if matchFlag == -1 {
			if c == '@' {
				start = i
				matchFlag = 1
			}
		} else if matchFlag == 1 {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '_')) {
				end = i
				break
			}
		}
	}

	if matchFlag == -1 {
		return 0, 0, false
	}
	if end > start && IsUsername(s[start+1:end]) {
		return start, end, true
	}

	//当start>0,判断@前一位是否为数字或字母,用于邮箱之类判断
	if start > 0 {
		c := s[start-1]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '_') {
			start = 0
		}
	}
	return start, end, false
}
