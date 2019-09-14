/*Copyright IBM Corporation 2019
 
 LICENSE: Apache License
          Version 2.0, January 2004
          http://www.apache.org/licenses/
          
 The following code is sample code created by IBM Corporation.
 This sample code is not part of any standard IBM product and
 is provided to you solely for the purpose of assisting you in
 the development of your applications.  The code is provided
 'as is', without warranty or condition of any kind.  IBM shall
 not be liable for any damages arising out of your use of the
 sample code, even if IBM has been advised of the possibility
 of such damages.
*/

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
	"errors"

	"github.com/magiconair/properties"
	"github.com/spf13/cobra"
)

var server string
var userID string
var password string

type servicesList struct {
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
		req, err := createRequest(server + "/services")
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			if resp.StatusCode == 403 || resp.StatusCode == 401 {
				return errors.New("Unable to authenticate with Db2")
			} 
			return errors.New(resp.Status)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var services servicesList
		json.Unmarshal(body, &services)
		for _, service := range services.DB2Services {
			req, err := createRequest(service.ServiceURL)
			if err != nil {
				return err
			}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				if resp.StatusCode == 403 || resp.StatusCode == 401 {
					return errors.New("Unable to authenticate with Db2")
				} 
				return errors.New(resp.Status)
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
			serviceProperties := createProperties(service)
			err = os.Mkdir(serviceName, 0775)
			if (err != nil) {
				return err
			}
			err = ioutil.WriteFile("./"+serviceName+"/"+serviceName+"Request.json", []byte(requestSchema), 0644)
			if (err != nil) {
				return err
			}
			err = ioutil.WriteFile("./"+serviceName+"/"+serviceName+"Response.json", []byte(responseSchema), 0644)
			if (err != nil) {
				return err
			}
			f, err := os.Create("./" + serviceName + "/service.properties")
			if (err != nil) {
				return err
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			_, err = serviceProperties.Write(w, properties.UTF8)
			if (err != nil) {
				return err
			}
			w.Flush()
		}
		return nil
	},
}

//Creates an HTTP request for calling Db2 on the specified url
func createRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(userID, password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

//Create the properties object for the given Db2 service
func createProperties(service map[string]interface{}) *properties.Properties {
	serviceName := service["serviceName"].(string)
	p := properties.NewProperties()
	p.Set("name", serviceName)
	p.Set("provider", "rest")
	p.Set("version", "1.0")
	p.Set("connectionRef", "db2Connection")
	p.Set("description", service["serviceDescription"].(string))
	p.Set("requestSchemaFile", "./"+serviceName+"Request.json")
	p.Set("responseSchemaFile", "./"+serviceName+"Response.json")
	p.Set("verb", "POST")
	u, _ := url.Parse(service["serviceURL"].(string))
	p.Set("uri", u.RequestURI())
	return p
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
