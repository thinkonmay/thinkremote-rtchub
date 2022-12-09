package system

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/jaypipes/ghw"
	"github.com/pion/stun"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

// SysInfo saves the basic system information
type SysInfo struct {
	Hostname string   `json:"os"`
	CPU      string   `json:"cpu"`
	RAM      string   `json:"ram"`
	Bios     string   `json:"bios"`
	Gpu      []string `json:"gpus"`
	Disk     []string `json:"disks"`
	Network  []string `json:"networks"`
    IP       string   `json:"ip"`
    PrivateIP string   `json:"privateip"`
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

func GetPublicIP() string {
	result := ""
	addr := "stun.l.google.com:19302"

	// we only try the first address, so restrict ourselves to IPv4
	c, err := stun.Dial("udp4", addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
		if res.Error != nil {
			log.Fatalln(res.Error)
		}
		var xorAddr stun.XORMappedAddress
		if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
			log.Fatalln(getErr)
		}
		result = xorAddr.IP.String()
	}); err != nil {
		log.Fatal("do:", err)
	}
	if err := c.Close(); err != nil {
		log.Fatalln(err)
	}


	return result;
}

func GetInfor() *SysInfo {
	hostStat, _ := host.Info()
	vmStat, _ := mem.VirtualMemory()

	gpu, err := ghw.GPU()
	bios, err := ghw.BIOS()
	pcies, err := ghw.Block()
	cpus, err := ghw.CPU()
	networks, err := ghw.Network()
    if err != nil  {
        fmt.Printf("unable to get information from system: %s",err.Error())
        return nil; 
    }

	ret := &SysInfo{
		CPU:      cpus.Processors[0].Model,
		RAM:      fmt.Sprintf("%dMb", vmStat.Total/1024/1024),
		Bios:     fmt.Sprintf("%s version %s",bios.Vendor,bios.Version),
	}

    if hostStat.VirtualizationSystem == "" {
		ret.Hostname = fmt.Sprintf("Baremetal %s ( OS %s %s) (arch %s)", hostStat.Hostname, hostStat.Platform,hostStat.PlatformVersion,hostStat.KernelArch)
    } else {
		ret.Hostname = fmt.Sprintf("VM %s ( OS %s %s) (arch %s)", hostStat.Hostname, hostStat.Platform,hostStat.PlatformVersion,hostStat.KernelArch)

    }

	for _, i := range gpu.GraphicsCards {
		ret.Gpu = append(ret.Gpu, i.DeviceInfo.Product.Name)
	}
	for _, i := range pcies.Disks {
		ret.Disk = append(ret.Disk, fmt.Sprintf("%s (Size %dGb)", i.Model, i.SizeBytes/1024/1024/1024))
	}
	for _, i := range networks.NICs {
		if (i.MacAddress != "") {
			ret.Network = append(ret.Network , fmt.Sprintf("%s (MAC Address %s)",i.Name,i.MacAddress));
		} else {
			ret.Network = append(ret.Network , i.Name);
		}
	}


	// Get preferred outbound ip of this machine
	ret.IP = GetPublicIP()
	ret.PrivateIP = GetOutboundIP().String()

    buf := make([]byte,50);
    resp,err := http.Get("https://api.ipify.org");
	if err != nil {
        ret.IP = "unavailable";
    } else if resp.StatusCode != 200{
        ret.IP = "unavailable";
	} else {
        size,_ := resp.Body.Read(buf)
        ret.IP = string(buf[:size]);

    }


	return ret
}



