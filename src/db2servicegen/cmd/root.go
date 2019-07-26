package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/magiconair/properties"
	"github.com/spf13/cobra"
)

var server string
var userID string
var password string

/*ServicesList The list of all the services in Db2*/
type ServicesList struct {
	DB2Services []struct {
		ServiceName         string      `json:"ServiceName"`
		ServiceCollectionID interface{} `json:"ServiceCollectionID"`
		ServiceDescription  string      `json:"ServiceDescription"`
		ServiceProvider     string      `json:"ServiceProvider"`
		ServiceURL          string      `json:"ServiceURL"`
	} `json:"DB2Services"`
}

var rootCmd = &cobra.Command{
	Use:   "db2servicegen",
	Short: "Create z/OS Connect EE service projects",
	Long:  "Query Db2 for available services and generate REST Client SP service projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", server + "/services", nil)
		req.SetBasicAuth(userID, password)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var services ServicesList
		json.Unmarshal(body, &services)
		for _, service := range services.DB2Services {
			// fmt.Printf("%s\n", service.ServiceName)
			req, _ := http.NewRequest("GET", service.ServiceURL, nil)
			req.SetBasicAuth(userID, password)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			var serviceInfo map[string]interface{}
			json.Unmarshal(body, &serviceInfo)
			service := serviceInfo[service.ServiceName].(map[string]interface{})
			serviceName := service["serviceName"].(string)
			fmt.Printf("Name: %s\tDescription: %s\n", service["serviceName"].(string), service["serviceDescription"].(string))
			request, _ := json.Marshal(service["RequestSchema"].(map[string]interface{}))
			requestSchema := strings.ReplaceAll(string(request), "\"null\",", "")
			response, _ := json.Marshal(service["ResponseSchema"].(map[string]interface{}))
			responseSchema := strings.ReplaceAll(string(response), "\"null\",", "")
			os.Mkdir(serviceName, 0775)
			ioutil.WriteFile("./"+serviceName+"/"+serviceName+"Request.json", []byte(requestSchema), 0644)
			ioutil.WriteFile("./"+serviceName+"/"+serviceName+"Response.json", []byte(responseSchema), 0644)
			p := properties.NewProperties()
			p.Set("name", serviceName)
			p.Set("provider", "rest")
			p.Set("version", "1.0")
			p.Set("connectionRef", "db2Connection")
			p.Set("description", service["serviceName"].(string))
			p.Set("requestSchemaFile", "./"+serviceName+"Request.json")
			p.Set("responseSchemaFile", "./"+serviceName+"Response.json")
			p.Set("verb", "POST")
			u, _ := url.Parse(service["serviceName"].(string))
			p.Set("uri", u.RequestURI())
			f, _ := os.Create("./" + serviceName + "/service.properties")
			defer f.Close()
			w := bufio.NewWriter(f)
			p.Write(w, properties.UTF8)
			w.Flush()
		}
		return nil
	},
}

//Execute runs the command with no sub-command. Sets up the parameters and handles any errors thrown.
func Execute() {
	rootCmd.Flags().StringVarP(&server, "server", "s", "", "The hostname and port of the Db2 server")
	rootCmd.MarkFlagRequired("server")
	rootCmd.Flags().StringVarP(&userID, "user", "u", "", "The userId for accessing the Db2 services")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "The password for accessing the Db2 services")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
