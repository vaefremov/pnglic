package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                 int    `yaml:"port"`
	PublicName           string `yaml:"serverPublicName"`
	DSN                  string `yaml:"dsn"`
	LicfileEncoderLegacy string `yaml:"encoderOld"`
	LicfileEncoderV3     string `yaml:"encoderV3"`
	SecretsHasp          string `yaml:"secretsHASP"`
	SecretsGuardant      string `yaml:"secretsGuardant"`
	StaticContent        string `yaml:"static"`
	AdminName            string `yaml:"adminName"`
	AdminPass            string `yaml:"adminPass"`
	AdminMail            string `yaml:"adminMail"`
	MailServer           string `yaml:"mailServer"`
	MailPort             int    `yaml:"mailPort"`
	MailUser             string `yaml:"mailUser"`
	MailPass             string `yaml:"mailPass"`
	BackMail             string `yaml:"backMail"`
	DaysToExpire1        int    `yaml:"daysToExpire1"`
	DaysToExpire2        int    `yaml:"daysToExpire2"`
}

const (
	defaultPort               = 9995
	defaultDsn                = "/Users/efremov/Projects/LIC/PNGLicenseManager/Backend/licset.sqlite"
	defaultLmGenPath          = "/Users/efremov/Projects/LIC/LmgenEmul/lmgen_hasp.sh"
	defaultSecretPathHASP     = "/Users/efremov/Projects/LIC/lm/licenses/5A6DD26A.secret"
	defaultSecretPathGuardant = "/Users/efremov/Projects/LIC/lm/licenses/5A6DD26A.secret"
	defaultStaticContent      = "templates"
	defaultAdminName          = "admin"
	defaultAdminPass          = "admin"
	defaultAdminMail          = "admin@pangea.ru"
	defaultMailPort           = 25
	defaultDaysToExpire1      = 7
	defaultDaysToExpire2      = 1
	defaultBackMail           = ""
	defaultPublicName         = "localhost"
)

var port = flag.Int("p", -1, "Port to start server on")
var dsnPtr = flag.String("dsn", "", "Path to sqlite3 database")
var lmgenPath = flag.String("lmgenHasp", "", "Path to legacy version of lmgen")
var secretPathHASP = flag.String("sHasp", "", "Path to the HASP secret file")
var secretPathGuardant = flag.String("sGuardant", "", "Path to the Guardant secret file")
var staticFilesPath = flag.String("static", "", "Path to static files and templates")
var adminName = flag.String("admin", "", "Name of admin user")
var adminPass = flag.String("password", "", "Password of admin user")
var adminMail = flag.String("eMail", "", "E-Mail of admin user")
var backMail = flag.String("backMail", "", "2nd mail address for messages")
var publicName = flag.String("publicName", "", "server public name")

// NewConfig decodes config from the specified JSON and updates it with the CLI values
func NewConfig(configPath string) (conf *Config) {
	conf = &Config{}
	conf.InsertDefaults()
	log.Printf("Reading config from: %s", configPath)
	f, err := os.Open(configPath)
	defer f.Close()
	if err != nil {
		log.Println("Warning", err)
		return
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(conf)
	if err != nil {
		log.Println("Warning", err)
	}
	conf.UpdateFromCLI()
	return
}

// Report prints the effective config parameters
func (c Config) Report() {
	fmt.Printf("Effective config:\n")
	fmt.Printf("  DSN:                   %s\n", c.DSN)
	fmt.Printf("  Port:                  %d\n", c.Port)
	fmt.Printf("  Server public name:    %s\n", c.PublicName)
	fmt.Printf("  Legacy encoder:        %s\n", c.LicfileEncoderLegacy)
	fmt.Printf("  V3 encoder:            %s\n", c.LicfileEncoderV3)
	fmt.Printf("  HASP secrets file:     %s\n", c.SecretsHasp)
	fmt.Printf("  Guardant secrets file: %s\n", c.SecretsGuardant)
	fmt.Printf("  Static files in:       %s\n", c.StaticContent)
	fmt.Printf("  Admin name is:         %s\n", c.AdminName)
	fmt.Printf("  Admin mail is:         %s\n", c.AdminMail)
	fmt.Printf("  Mail server:           %s:%d\n", c.MailServer, c.MailPort)
	fmt.Printf("  Mail user:             %s\n", c.MailUser)
	fmt.Printf("  Back mail:             %s\n", c.BackMail)
	fmt.Printf("  Exp. Term 1:           %d\n", c.DaysToExpire1)
	fmt.Printf("  Exp. Term 2:           %d\n", c.DaysToExpire2)
}

func (c Config) Write(configPath string) (err error) {
	if _, err = os.Stat(configPath); err == nil {
		return fmt.Errorf("file %s exists, will not write into it", configPath)
	}
	if f, err := os.Create(configPath); err == nil {
		encoder := yaml.NewEncoder(f)
		err = encoder.Encode(c)
		if err == nil {
			fmt.Printf("New config written to %s. \n", configPath)
		}
		return err
	}
	return
}

// UpdateFromCLI inserts the values of config parameters specified via the CLI
func (c *Config) UpdateFromCLI() {
	if *dsnPtr != "" {
		c.DSN = *dsnPtr
	}
	if *port != -1 {
		c.Port = *port
	}
	if *lmgenPath != "" {
		c.LicfileEncoderLegacy = *lmgenPath
	}
	if *secretPathHASP != "" {
		c.SecretsHasp = *secretPathHASP
	}
	if *secretPathGuardant != "" {
		c.SecretsGuardant = *secretPathGuardant
	}
	if *staticFilesPath != "" {
		c.StaticContent = *staticFilesPath
	}
	if *adminName != "" {
		c.AdminName = *adminName
	}
	if *adminPass != "" {
		c.AdminPass = *adminPass
	}
	if *adminMail != "" {
		c.AdminMail = *adminMail
	}
	if *backMail != "" {
		c.BackMail = *backMail
	}
	if *publicName != "" {
		c.PublicName = *publicName
	}
}

// InsertDefaults inserts the built-in default values of config parameters
func (c *Config) InsertDefaults() {
	c.DSN = defaultDsn
	c.Port = defaultPort
	c.LicfileEncoderLegacy = defaultLmGenPath
	c.SecretsHasp = defaultSecretPathHASP
	c.SecretsGuardant = defaultSecretPathGuardant
	c.StaticContent = defaultStaticContent
	c.AdminName = defaultAdminName
	c.AdminPass = defaultAdminPass
	c.AdminMail = defaultAdminMail
	c.MailPort = defaultMailPort
	c.DaysToExpire1 = defaultDaysToExpire1
	c.DaysToExpire2 = defaultDaysToExpire2
	c.BackMail = defaultBackMail
	c.PublicName = defaultPublicName
}
