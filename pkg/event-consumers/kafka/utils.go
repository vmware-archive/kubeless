/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"strconv"

	"github.com/sirupsen/logrus"
)

// GetTLSConfiguration build TLS configuration for kafka
func GetTLSConfiguration(caFile string, certFile string, keyFile string, insecure string) (*tls.Config, bool, error) {
	logrus.Debugf("configure tls %s %s %s %b", caFile, certFile, keyFile, insecure)
	isInsecure, _ := strconv.ParseBool(insecure)
	if (caFile == "" && (certFile == "" || keyFile == "")) && !isInsecure {
		return nil, false, nil
	}
	t := &tls.Config{}
	if caFile != "" {
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, false, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		t.RootCAs = caCertPool
	}

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, false, err
		}
		t.Certificates = []tls.Certificate{cert}
	}

	t.InsecureSkipVerify = isInsecure
	logrus.Debugf("TLS config %+v", t)

	return t, true, nil
}

// GetSASLConfiguration build SASL configuration for kafka
func GetSASLConfiguration(username string, password string) (string, string, bool) {
	if username != "" && password != "" {
		return username, password, true
	}
	return "", "", false
}
