package main

import (
  "bufio"
  "fmt"
  "io/ioutil"
  "encoding/json"
  "log"
  "os"
  "net/http"
  "net/url"
  "strings"

  "github.com/urfave/cli"
  "github.com/magiconair/properties"
)

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

var app = cli.NewApp()

func main() {
  var server string
  var userID string
  var password string

  app.Name = "db2servicegen"
  app.Usage = "Generate service projects for Db2 REST Services"
  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "server, s",
      Usage: "The Db2 server hostname and port",
      Destination: &server,
    },
    cli.StringFlag{
      Name: "user, u",
      Usage: "The User ID for accessing the server",
      Value: "",
      Destination: &userID,
    },
    cli.StringFlag{
      Name: "password, p",
      Usage: "The password for accessing the server",
      Value: "",
      Destination: &password,
    },
  }
  app.Action = func(c *cli.Context) error {
    client := &http.Client{}
    req, _ := http.NewRequest("GET", server, nil)
    req.SetBasicAuth(userID, password)
    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Type", "application/json")
    resp, err := client.Do(req)
    if(err != nil){
      return cli.NewExitError(err, 1)
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
      if(err != nil){
        return cli.NewExitError(err, 1)
      }
      defer resp.Body.Close()
      body, _ := ioutil.ReadAll(resp.Body)
      var serviceInfo map[string]interface{}
      json.Unmarshal(body, &serviceInfo)
      service := serviceInfo[service.ServiceName].(map[string]interface{})
      serviceName := service["serviceName"].(string)
      fmt.Printf("Name: %s\tDescription: %s\n", service["serviceName"].(string), service["serviceDescription"].(string))
      request, _ := json.Marshal(service["RequestSchema"].(map[string]interface{}));
      requestSchema := strings.ReplaceAll(string(request), "\"null\",", "")
      response, _ := json.Marshal(service["ResponseSchema"].(map[string]interface{}));
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
      f, _ := os.Create("./"+serviceName+"/service.properties")
      defer f.Close()
      w := bufio.NewWriter(f)
      p.Write(w, properties.UTF8)
      w.Flush()
    }
    return nil
  }
  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}