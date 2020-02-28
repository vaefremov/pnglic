package xmlutils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vaefremov/pnglic/pkg/xmlutils"
)

func TestParseLicenseFileXML(t *testing.T) {
	serv, res, err := xmlutils.ParseLicenseFileXML([]byte(licenseFile1))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(serv, res)
	// t.Error("debug", err)
}

func TestReorderSingleFeaturesFirst(t *testing.T) {
	res, err := xmlutils.ReorderSingleFeaturesFirst(licenseFile1)
	if err != nil {
		t.Error("Error returned by parser", err)
	}
	assert.Equal(t, licenseFile1Sorted, res)
	fmt.Print(string(res))
	// t.Error(nil)
}

// Test data

const licenseFile1 = `
<?xml version="1.0"?>
<!DOCTYPE license_server>

<license_server port="1234" id="74B6BCC0">

<package
        id="CT_INTERFACE"
        version="1.00"
        start="20.06.2018"
        end="12.01.2021"
        count="1"
        dupgroup="DISP"
        code="F5B9B58295A4EA6E740F672F9FF557D3">
    <feature id="MAPCENTER"/>
    <feature id="CT_CLASSI"/>
    <feature id="CT_CLUSTER"/>
    <feature id="CT_COKRIGING"/>
</package>
<feature
        id="LM_CONSOLE"
        version="1.00"
        start="01.01.2018"
        end="12.01.2021"
        count="1"
        dupgroup="DISP"
        code="211A7BD8D3025B9B09ECC47B919B767F"/>
<package
        id="RV_PROCESSES"
        version="1.00"
        start="20.06.2018"
        end="12.01.2021"
        count="1"
        dupgroup=""
        code="5FF8C4C048FA0DA0A690B438089C8F7A">
    <feature id="RV_2_DERIVATIVE"/>
    <feature id="RV_AMPL_EQUAL"/>
    <feature id="RV_ANISOTR"/>
</package>
<feature
        id="MULTILOG"
        version="1.00"
        start="20.06.2018"
        end="12.01.2021"
        count="2"
        dupgroup="DISP"
        code="5B80BBF68DBADA470BEFCC6818E98A4A"/>
</license_server>
`
const licenseFile1Sorted = `<?xml version="1.0"?><!DOCTYPE license_server>

<license_server port="1234" id="74B6BCC0">
    <feature
            id="LM_CONSOLE"
            version="1.00"
            start="01.01.2018"
            end="12.01.2021"
            count="1"
            dupgroup="DISP"
            code="211A7BD8D3025B9B09ECC47B919B767F" >
    </feature>
    <feature
            id="MULTILOG"
            version="1.00"
            start="20.06.2018"
            end="12.01.2021"
            count="2"
            dupgroup="DISP"
            code="5B80BBF68DBADA470BEFCC6818E98A4A" >
    </feature>
    <package
            id="CT_INTERFACE"
            version="1.00"
            start="20.06.2018"
            end="12.01.2021"
            count="1"
            dupgroup="DISP"
            code="F5B9B58295A4EA6E740F672F9FF557D3" >
        <feature id="MAPCENTER" />
        <feature id="CT_CLASSI" />
        <feature id="CT_CLUSTER" />
        <feature id="CT_COKRIGING" />
    </package>
    <package
            id="RV_PROCESSES"
            version="1.00"
            start="20.06.2018"
            end="12.01.2021"
            count="1"
            dupgroup=""
            code="5FF8C4C048FA0DA0A690B438089C8F7A" >
        <feature id="RV_2_DERIVATIVE" />
        <feature id="RV_AMPL_EQUAL" />
        <feature id="RV_ANISOTR" />
    </package>
</license_server>
`

const licenseFileSimple = `
<?xml version="1.0"?>
<!DOCTYPE license_server>

<license_server port="1234" id="74B6BCC0">

<package
        id="CT_INTERFACE"
        version="1.00"
        start="20.06.2018"
        end="12.01.2021"
        count="1"
        dupgroup="DISP"
        code="F5B9B58295A4EA6E740F672F9FF557D3">
    <feature id="MAPCENTER"/>
    <feature id="CT_CLASSI"/>
    <feature id="CT_CLUSTER"/>
    <feature id="CT_COKRIGING"/>
</package>
<package
        id="RV_PROCESSES"
        version="1.00"
        start="20.06.2018"
        end="12.01.2021"
        count="1"
        dupgroup=""
        code="5FF8C4C048FA0DA0A690B438089C8F7A">
    <feature id="RV_2_DERIVATIVE"/>
    <feature id="RV_AMPL_EQUAL"/>
    <feature id="RV_ANISOTR"/>
</package>
</license_server>
`
