package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"gopkg.in/ini.v1"
)

type PipOp struct {
	Account  string
	Region   string
	PublicIp string
}

func main() {

	regions := []string{"eu-central-1", "ap-southeast-1", "eu-west-1"}
	profiles := GetLocalAwsProfiles()

	pipOps := getAppPips(profiles, regions)

	data, err := json.Marshal(pipOps)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("------------------------------------------output---------------------------------------------------")
	fmt.Println(string(data))
	fmt.Println("------------------------------------------output---------------------------------------------------")

	// ioutil.WriteFile("./ipsa.json", data, 0644)

}

func getAppPips(profiles []string, regions []string) []PipOp {
	pipOps := make([]PipOp, 0)

	var mx sync.Mutex

	var wg sync.WaitGroup
	for _, region := range regions {
		for _, profile := range profiles {
			wg.Add(1)
			go func(profile string, region string) {
				defer wg.Done()
				var dat map[string]interface{}
				str := getIpsinProfile(profile, region)
				if str == "" {
					return
				}

				if err := json.Unmarshal([]byte(str), &dat); err != nil {
					fmt.Println("1")
					panic(err)
				}

				results := dat["Results"].([]interface{})
				if len(results) > 0 {

					for _, result := range results {

						dt := result.(string)

						var dat2 map[string]interface{}

						if err := json.Unmarshal([]byte(dt), &dat2); err != nil {
							fmt.Println("2")
							panic(err)
						}

						configuration := dat2["configuration"].(map[string]interface{})
						association := configuration["association"].(map[string]interface{})
						pip := association["publicIp"].(string)
						mx.Lock()
						pipOps = append(pipOps, PipOp{Account: profile, Region: region, PublicIp: pip})
						mx.Unlock()
					}
				}
			}(profile, region)
		}
	}
	wg.Wait()
	return pipOps
}

const ShellToUse = "bash"

func getIpsinProfile(profile string, region string) string {

	cmdStr := fmt.Sprintf("aws configservice select-resource-config --expression \"SELECT resourceId, resourceName, resourceType, configuration.association.publicIp,  availabilityZone,  awsRegion WHERE  resourceType='AWS::EC2::NetworkInterface'  AND configuration.association.publicIp>'0.0.0.0'\" --profile %v --region %v", profile, region)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", cmdStr)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		if strings.Contains(stderr.String(), "ExpiredTokenException") {
			fmt.Printf("Toekn expited for the profile %v, region %v. Thus skipping \n", profile, region)
			return ""
		}
		fmt.Println(cmd.Stderr)
		fmt.Println(profile + region)
		panic(err)
	}
	return stdout.String()
}

func GetLocalAwsProfiles() []string {
	fname := config.DefaultSharedCredentialsFilename()
	f, err := ini.Load(fname)
	if err != nil {
	} else {
		arr := []string{}
		for _, v := range f.Sections() {
			if len(v.Keys()) != 0 {
				arr = append(arr, v.Name())
			}
		}
		return arr
	}
	return []string{}
}
