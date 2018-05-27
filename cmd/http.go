package cmd

import (
	"fmt"
	"strings"
	"os"
	"bytes"
	"io"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"github.com/buger/jsonparser"
	"encoding/json"
)

func httpCheck(args []string) []string {
	resourceUrls := []string{}
	if len(args) == 0{
		fmt.Println("Error: Must set at least one resource url")
		os.Exit(0)
	}else{
		for i := 0; i < len(args); i++ {
			resourceUrls = append(resourceUrls,args[i])

		}
	}
	return resourceUrls
}

func createFile(filepath string, body []byte) error{
	out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(body))
	if err != nil {
        return err
    }
    return nil
}

func downloadFileData(url string) string{
	newUrl :=  strings.Replace(url,"?download","",1)
	downloadData := connect("GET",newUrl,nil)
	fileName, err := jsonparser.GetString(downloadData,"fileName")
	if err != nil{
		fmt.Println(err)
	}
	return fileName
}

func makeRequest(verb string, url string, optionalJson io.Reader, ch chan <-[]byte) {
	byteData := connect(verb,url,optionalJson)
	m, ok := gjson.Parse(string(byteData)).Value().(map[string]interface{})
	jsonData := []byte{}

	if !ok && strings.Index(url,"?download") != -1{
		fileName := downloadFileData(url)
		createFile(fileName,byteData)
	}else if !ok{
        fmt.Println("Error")
    }else{
		formattedJson, _ := json.MarshalIndent(m,"","  ")
		jsonData = formattedJson
    }

	ch <- jsonData

}

func runHttp(cmd *cobra.Command, args []string) error {

	resourceUrls := httpCheck(args)
	resourceUrlsCount := len(resourceUrls) 
	httpVerb := strings.ToUpper(cmd.Use)

	ch := make(chan []byte)

	jsonData := bytes.NewReader([]byte(data))
	
	for i := 0; i < resourceUrlsCount; i++ {
		go makeRequest(httpVerb,resourceUrls[i], jsonData, ch)
		
		fmt.Fprintf(os.Stdout, "%s", <-ch)
		if httpVerb!="GET"{
			return nil;
		}
	}

	return nil
}

var get = &cobra.Command{
	Use: "get",
	Short: "Performs one or more GET requests",
	Long: "\033[93mPerforms one or more GET requests and returns parsed results\033[0m \033[0;32m\n\nSingle Query Example: \033[0m \n$ osvc-rest query \"DESCRIBE\" -u $OSC_ADMIN -p $OSC_PASSWORD -i $OSC_SITE \033[0;32m\n\nMultiple Queries Example:\033[0m \n$ osvc-rest query \"SELECT * FROM INCIDENTS LIMIT 100\" \"SELECT * FROM SERVICEPRODUCTS LIMIT 100\" \"SELECT * FROM SERVICECATEGORIES LIMIT 100\" -u $OSC_ADMIN -p $OSC_PASSWORD -i $OSC_SITE",
	RunE: runHttp,
}

var data string

func checkPostPatchFlags(flags *pflag.FlagSet) error {

	if data == "" {
		fmt.Println("\033[31mError: Must send JSON Data for POST and PATCH requests; use the --data flag")
		os.Exit(0)
	}

	return nil
}

var post = &cobra.Command{
	Use: "post",
	Short: "Performs a POST request",
	Long: "\033[93mPerforms a POST request and returns parsed results\033[0m \033[0;32m\n\nExample: \033[0m \n$ osvc-rest post \"opportunities\" --data '{\"name\":\"PCS- 100 laptops\"}' -u $OSC_ADMIN -p $OSC_PASSWORD -i $OSC_SITE\n\n",
	PreRunE:func(cmd *cobra.Command, args []string) error {		
		return checkPostPatchFlags(cmd.Flags())
	},
	RunE: runHttp,
}

var patch = &cobra.Command{
	Use: "patch",
	Short: "Performs a PATCH request",
	Long: "\033[93mPerforms a PATCH request; if successful, nothing is returned \033[0m \033[0;32m\n\nExample: \033[0m \n$ osvc-rest patch \"opportunities/1\" --data '{\"name\":\"PCS- 100 laptops UPDATED\"}' -u $OSC_ADMIN -p $OSC_PASSWORD -i $OSC_SITE\n\n",
	PreRunE:func(cmd *cobra.Command, args []string) error {		
		return checkPostPatchFlags(cmd.Flags())
	},
	RunE: runHttp,
}

var delete = &cobra.Command{
	Use: "delete",
	Short: "Performs a DELETE request",
	Long: "\033[93mPerforms a DELETE request; if successful, nothing is returned \033[0m \033[0;32m\n\nExample: \033[0m \n$ osvc-rest delete \"opportunities/1\" -u $OSC_ADMIN -p $OSC_PASSWORD -i $OSC_SITE\n\n",
	RunE: runHttp,
}

func init(){
	RootCmd.AddCommand(get)
	post.Flags().StringVarP(&data,"data","j","", "Sets the JSON data to be sent for the POST request")
	patch.Flags().StringVarP(&data,"data","j","", "Sets the JSON data to be sent for the POST request")
	RootCmd.AddCommand(post)
	RootCmd.AddCommand(patch)
	RootCmd.AddCommand(delete)
}