// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datax

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
)

const (
	fieldStringMaxLen = 102400
)

var (
	timeFormatMap = []string{
		"Y", "2006",
		"y", "06",
		"m", "01",
		"d", "02",
		"H", "15",
		"h", "03",
		"i", "04",
		"s", "05",
		"date", "2006-01-02",
		"datetime", "2006-01-02 15:04:05",
		"atom", time.RFC3339,
	}
	timeFormator = strings.NewReplacer(timeFormatMap...)
	s2Replacer   *strings.Replacer
)

var (
	regElement    = regexp.MustCompile("\\<[\\S\\s]+?\\>")
	regStyle      = regexp.MustCompile("\\<style[\\S\\s]+?\\</style\\>")
	regScript     = regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
	regMultiLine  = regexp.MustCompile("\\n\\n+")
	regMultiSpace = regexp.MustCompile("\\s{2,}")
	regLineSpace  = regexp.MustCompile("\\n\\s*\\n")
	regMath       = regexp.MustCompile("\\$\\$(.*?)\\$\\$")
	mkp           = bluemonday.UGCPolicy()
	htmlp         = bluemonday.UGCPolicy()
	shtmlp        = bluemonday.UGCPolicy()
)

func init() {
	mkp.AllowAttrs("class").OnElements("code")

	//
	shtmlp.AllowElements("script")
	shtmlp.AllowElements("button", "style")

	//
	shtmlp.AllowAttrs("class").OnElements("div")
	shtmlp.AllowAttrs("style").OnElements("div")

	//
	shtmlp.AllowAttrs("class").OnElements("button")
	shtmlp.AllowAttrs("onclick").OnElements("button")

	//
	shtmlp.AllowAttrs("href").OnElements("a")
	shtmlp.AllowAttrs("target").OnElements("a")
	shtmlp.AllowAttrs("class").OnElements("a")
	shtmlp.AllowAttrs("class").OnElements("img")
	shtmlp.AllowAttrs("class").OnElements("span")
}

func s2_replace(s string) string {
	if s2Replacer == nil {
		s2Replacer = strings.NewReplacer(
			"{{hp_storage_service_endpoint}}", config.SysConfigList.FetchString("storage_service_endpoint"),
		)
	}
	return s2Replacer.Replace(s)
}

func UnixtimeFormat(timeValue interface{}, formatTo string) string {

	var tp time.Time

	switch timeValue.(type) {
	case uint32:
		tp = time.Unix(int64(timeValue.(uint32)), 0)

	case int64:
		tp = time.Unix(timeValue.(int64), 0)

	default:
		tp = time.Now()
	}

	return tp.Format(timeFormator.Replace(formatTo))
}

func TimeFormat(timeString, formatFrom, formatTo string) string {

	tp, err := time.ParseInLocation(timeFormator.Replace(formatFrom), timeString, time.Local)
	if err != nil {
		return timeString
	}

	return tp.Format(timeFormator.Replace(formatTo))
}

func FieldTimeFormat(fields []*api.NodeField, colname, format string) string {

	val, _ := fieldValue(fields, colname)

	tp, err := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local)
	if err != nil {
		return val
	}

	format = timeFormator.Replace(format)

	return tp.Format(format)
}

func TextHtml2Str(src string) string {

	src = regElement.ReplaceAllStringFunc(src, strings.ToLower)

	src = regStyle.ReplaceAllString(src, "")
	src = regScript.ReplaceAllString(src, "")

	src = regElement.ReplaceAllString(src, "\n")
	src = regMultiSpace.ReplaceAllString(src, "\n")

	src = regMultiLine.ReplaceAllString(src, "\n\n")

	return strings.TrimSpace(src)
}

func fieldValue(fields []*api.NodeField, colname string) (string, map[string]string) {

	var (
		val   = ""
		attrs = map[string]string{}
	)

	for _, field := range fields {

		if field.Name != colname {
			continue
		}

		for _, v := range field.Attrs {
			attrs[v.Key] = v.Value
		}

		val = field.Value

		break
	}

	return val, attrs
}

func fieldValueCache(fields []*api.NodeField, colname string, cache_key string) (string, map[string]string, bool) {

	var (
		val    = ""
		attrs  = map[string]string{}
		cached = false
	)

	for _, field := range fields {

		if field.Name != colname {
			continue
		}

		for _, v := range field.Attrs {
			attrs[v.Key] = v.Value
		}

		if v := field.Caches.Get(cache_key); v != nil && len(v) > 0 {
			val, cached = v.String(), true
		} else {
			val = field.Value
		}

		break
	}

	return val, attrs, cached
}

func fieldValueCacheSet(fields []*api.NodeField, colname, value, cache_key string) {
	for _, field := range fields {

		if field.Name == colname {
			field.Caches.Set(cache_key, value)
			break
		}
	}
}

func FieldString(fields []*api.NodeField, colname string) string {

	val, _ := fieldValue(fields, colname)

	return val
}

func FieldSubString(fields []*api.NodeField, colname string, length int) string {

	if length < 1 {
		length = fieldStringMaxLen
	}

	val, _, cached := fieldValueCache(fields, colname, fmt.Sprintf("FieldSubString_%d", length))
	if cached {
		return val
	}

	ustr := []rune(val)

	if len(ustr) > length {
		val = string(ustr[0:length]) + "..."
	}

	fieldValueCacheSet(fields, colname, val, fmt.Sprintf("FieldSubString_%d", length))

	return val
}

func FieldDebug(fields []*api.NodeField, colname string, length int) template.HTML {

	val, attrs := fieldValue(fields, colname)

	if v, ok := attrs["format"]; ok {

		if v == "md" {
			unsafe := blackfriday.MarkdownCommon([]byte(val))
			val = string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
			v = "html"
		}

		if v == "html" || v == "shtml" {
			val = TextHtml2Str(val)
		}
	}

	if len(val) > length {
		val = val[:length] + "..."
	}

	return template.HTML(val)
}

func FieldHtml(fields []*api.NodeField, colname string) template.HTML {

	val, attrs, cached := fieldValueCache(fields, colname, "FieldHtml")
	if cached {
		return template.HTML(val)
	}

	fm, ok := attrs["format"]
	if !ok {
		fm = "text"
	}

	val = strings.TrimSpace(strings.Replace(val, "\r\n", "\n", -1))
	val = regMultiLine.ReplaceAllString(val, "\n\n")

	val = s2_replace(val)

	switch fm {

	case "md":

		if strings.Index(val, `$$`) > 0 {
			// val = regMath.ReplaceAllString(val, "<code class=\"language-math\">$1</code>")
			val = strings.Replace(val, `\\`, `\\\\`, -1)
		}

		unsafe := blackfriday.MarkdownCommon([]byte(val))
		val = string(mkp.SanitizeBytes(unsafe))

	case "html":
		val = htmlp.Sanitize(val)

	case "shtml":
		val = shtmlp.Sanitize(val)

	case "text":
		if lines := strings.Split(val, "\n\n"); len(lines) > 1 {
			val = "<p>" + strings.Join(lines, "</p><p>") + "</p>"
			val = strings.Replace(val, "\n", "<br>", -1)
		}
		fallthrough

	default:
		val = htmlp.Sanitize(val)
	}

	fieldValueCacheSet(fields, colname, val, "FieldHtml")

	return template.HTML(val)
}

func StringSub(s string, start, length int) string {

	bt := []rune(s)

	if start < 0 {
		start = 0
	}

	if length < 1 {
		length = 1
	}

	end := start + length

	if end >= len(bt) {
		end = len(bt)
	}

	if end <= start {
		return ""
	}

	return string(bt[start:end])
}

func FieldSubHtml(fields []*api.NodeField, colname string, length int) template.HTML {

	var field *api.NodeField
	for _, v := range fields {
		if v.Name == colname {
			field = v
			break
		}
	}
	if field == nil {
		return ""
	}

	if length < 1 {
		length = fieldStringMaxLen
	}

	var (
		cache_key = fmt.Sprintf("fhsp_%d", length)
	)

	if v := field.Caches.Get(cache_key); v != nil {
		return template.HTML(v.String())
	}

	val := field.Value

	fm := "text"
	if attr := field.Attrs.Get("format"); attr != nil {
		fm = attr.String()
	}

	if fm == "md" {
		unsafe := blackfriday.MarkdownCommon([]byte(val))
		val = string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
	}

	ustr := []rune(TextHtml2Str(val))

	if len(ustr) > length {
		val = string(ustr[0:length]) + "..."
	} else {
		val = string(ustr)
	}

	val = strings.Replace(val, "\n", "<br>", -1)

	if lines := strings.Split(val, "\n\n"); len(lines) > 1 {
		val = "<p>" + strings.Join(lines, "</p><p>") + "</p>"
	}

	field.Caches.Set(cache_key, val)

	return template.HTML(val)
}

func FieldHtmlSubPrint(node_entry api.Node, colname string, length int, lang string) template.HTML {

	var field *api.NodeField
	for _, v := range node_entry.Fields {
		if v.Name == colname {
			field = v
			break
		}
	}
	if field == nil {
		return ""
	}

	if length < 1 {
		length = fieldStringMaxLen
	}

	var (
		cache_key = fmt.Sprintf("fhsp_%d", length)
		val       string
	)

	if field.Langs != nil {

		if lang := field.Caches.Get(cache_key + lang); lang != nil {
			return template.HTML(lang.String())
		}

		if lang := field.Langs.Items.Get(lang); lang != nil {
			val = lang.String()
		}
	}

	if val == "" {
		if v := field.Caches.Get(cache_key); v != nil {
			return template.HTML(v.String())
		}
		val = field.Value
		lang = ""
	}

	fm := "text"
	if attr := field.Attrs.Get("format"); attr != nil {
		fm = attr.String()
	}

	if fm == "md" {
		unsafe := blackfriday.MarkdownCommon([]byte(val))
		val = string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
	}

	ustr := []rune(TextHtml2Str(val))

	if len(ustr) > length {
		val = string(ustr[0:length]) + "..."
	} else {
		val = string(ustr)
	}

	val = strings.Replace(val, "\n", "<br>", -1)

	if lines := strings.Split(val, "\n\n"); len(lines) > 1 {
		val = "<p>" + strings.Join(lines, "</p><p>") + "</p>"
	}

	field.Caches.Set(cache_key+lang, val)

	return template.HTML(val)
}

func FieldStringPrint(node_entry api.Node, colname, lang string) string {
	for _, field := range node_entry.Fields {
		if field.Name == colname {
			if field.Langs != nil {
				if v := field.Langs.Items.Get(lang); v != nil {
					return v.String()
				}
			}
			return field.Value
		}
	}
	return ""
}

func FieldHtmlPrint(node_entry api.Node, colname, lang string) template.HTML {

	var field *api.NodeField
	for _, v := range node_entry.Fields {
		if v.Name == colname {
			field = v
			break
		}
	}
	if field == nil {
		return ""
	}

	var (
		cache_key = "fhp"
		val       string
	)

	if field.Langs != nil {

		if lang := field.Caches.Get(cache_key + lang); lang != nil {
			return template.HTML(lang.String())
		}

		if lang := field.Langs.Items.Get(lang); lang != nil {
			val = lang.String()
		}
	}

	if val == "" {
		if v := field.Caches.Get(cache_key); v != nil {
			return template.HTML(v.String())
		}
	}

	fm := "text"
	if attr := field.Attrs.Get("format"); attr != nil {
		fm = attr.String()
	}

	val = field_value_html_convert(fm, field.Value)
	field.Caches.Set(cache_key, val)

	if field.Langs != nil {
		for _, v := range field.Langs.Items {
			v.Value = field_value_html_convert(fm, v.Value)
			field.Caches.Set(cache_key+v.Key, v.Value)
			if v.Key == lang {
				val = v.Value
			}
		}
	}

	return template.HTML(val)
}

func field_value_html_convert(fm, val string) string {

	val = strings.TrimSpace(strings.Replace(val, "\r\n", "\n", -1))
	val = regMultiLine.ReplaceAllString(val, "\n\n")

	val = s2_replace(val)

	switch fm {

	case "md":

		if strings.Index(val, `$$`) > 0 {
			// val = regMath.ReplaceAllString(val, "<code class=\"language-math\">$1</code>")
			val = strings.Replace(val, `\\`, `\\\\`, -1)
		}

		unsafe := blackfriday.MarkdownCommon([]byte(val))
		val = string(mkp.SanitizeBytes(unsafe))

	case "html":
		val = htmlp.Sanitize(val)

	case "shtml":
		val = shtmlp.Sanitize(val)

	case "text":
		if lines := strings.Split(val, "\n\n"); len(lines) > 1 {
			val = "<p>" + strings.Join(lines, "</p><p>") + "</p>"
			val = strings.Replace(val, "\n", "<br>", -1)
		}
		fallthrough

	default:
		val = htmlp.Sanitize(val)
	}

	return val
}
