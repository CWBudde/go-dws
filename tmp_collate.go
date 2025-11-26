package main
import (
"fmt"
"golang.org/x/text/collate"
"golang.org/x/text/language"
)
func main(){
opts := [][]collate.Option{{collate.IgnoreCase},{collate.IgnoreDiacritics},{collate.IgnoreWidth},{collate.Loose},{collate.IgnoreCase, collate.Loose}}
words := []string{"cote","côte","coté","côté"}
for idx,opt := range opts{
 col:=collate.New(language.French, opt...)
 fmt.Println("opt",idx)
 for i:=0;i<len(words)-1;i++{
  for j:=i+1;j<len(words);j++{
   fmt.Println(words[i], words[j], col.CompareString(words[i], words[j]))
  }
 }
}
}
