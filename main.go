package main

import (
	branch "MaoServerDiscovery/cmd"
	"MaoServerDiscovery/util"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"os"
	"runtime"
)

var (
	silent bool
	//main_server_addr net.IP
	report_server_addr net.IP
	report_server_port uint32
	report_interval uint32
	web_server_addr net.IP
	web_server_port uint32
	dump_interval uint32
	refresh_interval uint32

	nat66Gateway bool
	nat66Persistent bool
	influxdbUrl string
	influxdbOrgBucket string
	influxdbToken string
)

var rootCmd = &cobra.Command{
	Use: "mao-service-discovery",
	Short:   "Mao-Service-Discovery, welcome to join our Github community. https://github.com/MaoJianwei/MaoServiceDiscovery",
	Long:    "Mao-Service-Discovery:\n\nService registry & Service discovery & Service keep-alive.\n\nWelcome to join our Github community. https://github.com/MaoJianwei/MaoServiceDiscovery",
	//Example: "beijing",
	Version: "1.0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			util.MaoLog(util.ERROR, fmt.Sprintf("Fail to execute rootCmd.Help(): %s", err))
		}
	},
}

var generalClientCmd = &cobra.Command{
	Use: "client",
	Short:   "Mao: Run general client. For common device/server.",
	//Long:    "Run general client. For common device/server.",
	//Example: "a",
	Version: "1.0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := readGeneralClientArgs(cmd); err != nil {
			util.MaoLog(util.ERROR, fmt.Sprintf("Wrong Args for general client: %s", err))
			return
		}

		if nat66Gateway == true {
			if runtime.GOOS != `linux` || os.Getgid() != 0 {
				util.MaoLog(util.ERROR, "nat66Gateway is usable only in linux with root privilege")
				return
			}
		}

		branch.RunGeneralClient(&report_server_addr, report_server_port, report_interval, silent,
			nat66Gateway, nat66Persistent, influxdbUrl, influxdbOrgBucket, influxdbToken)
	},
}

var serverCmd = &cobra.Command{
	Use: "server",
	Short:   "Mao: Run server.",
	//Long:    "Run general client. For common device/server.",
	//Example: "a",
	Version: "1.0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := readServerArgs(cmd); err != nil {
			util.MaoLog(util.ERROR, fmt.Sprintf("Wrong Args for server: %s", err))
			return
		}

		//ss,_ := rootCmd.PersistentFlags().GetString("report_server_addr")
		//report_server_addr = net.ParseIP(ss)
		//fmt.Printf("%v\n%v\n%v\n%v\n%v\n%v\n%v\n",
		//	report_server_addr,
		//	report_server_port,
		//	main_server_addr,
		//	web_server_addr,
		//	web_server_port,
		//	dump_interval,
		//	report_interval)
		//
		//fmt.Printf("---\n%v, %d\n", args, len(args))
		//return
		branch.RunServer(&report_server_addr, report_server_port, &web_server_addr, web_server_port, dump_interval, refresh_interval, silent)
	},
}

func init() {
	rootCmd.PersistentFlags().String("report_server_addr","::1","2001:db8::1")
	rootCmd.PersistentFlags().Uint32("report_server_port",28888,"28888")
	rootCmd.PersistentFlags().Bool("silent", false,"false")


	//serverCmd.Flags().String("main_server_addr","::","::")
	serverCmd.Flags().String("web_server_addr","::","::")
	serverCmd.Flags().Uint32("web_server_port",29999,"29999")

	serverCmd.Flags().Uint32("dump_interval", 1000, "1000")
	serverCmd.Flags().Uint32("refresh_interval", 1000, "1000")


	generalClientCmd.Flags().Uint32("report_interval", 1000, "1000")
	generalClientCmd.Flags().Bool("enable_aux_nat66_stat", false, "Usable only in linux with root privilege")

	generalClientCmd.Flags().Bool("enable_nat66_persistent", false, "Enable to upload stat to Influxdb")
	generalClientCmd.Flags().String("influxdb_url","","https://domain-or-ip:port")
	generalClientCmd.Flags().String("influxdb_org_bucket","","same name for Org and Bucket")
	generalClientCmd.Flags().String("influxdb_token","","token string from Influxdb")
}

func readRootArgs(cmd *cobra.Command) error {

	report_server_addr_str, err := rootCmd.PersistentFlags().GetString("report_server_addr")
	if err != nil {
		return err
	}
	report_server_addr = net.ParseIP(report_server_addr_str)
	if report_server_addr == nil {
		return errors.New("report_server_addr is invalid")
	}

	report_server_port, err = rootCmd.PersistentFlags().GetUint32("report_server_port")
	if err != nil {
		return err
	}
	if report_server_port < 1 || report_server_port > 65535 {
		return errors.New("report_server_port is invalid")
	}


	silent, err = rootCmd.PersistentFlags().GetBool("silent")
	if err != nil {
		return err
	}

	return nil
}

func readServerArgs(cmd *cobra.Command) error {

	if err := readRootArgs(cmd); err != nil {
		return err
	}

	//main_server_addr_str, err := cmd.Flags().GetString("main_server_addr")
	//if err != nil {
	//	return err
	//}
	//main_server_addr = net.ParseIP(main_server_addr_str)
	//if main_server_addr == nil {
	//	return errors.New("main_server_addr is invalid")
	//}

	web_server_addr_str, err := cmd.Flags().GetString("web_server_addr")
	if err != nil {
		return err
	}
	web_server_addr = net.ParseIP(web_server_addr_str)
	if web_server_addr == nil {
		return errors.New("web_server_addr is invalid")
	}

	web_server_port, err = cmd.Flags().GetUint32("web_server_port")
	if err != nil {
		return err
	}
	if web_server_port < 1 || web_server_port > 65535 {
		return errors.New("web_server_port is invalid")
	}

	dump_interval, err = cmd.Flags().GetUint32("dump_interval")
	if err != nil {
		return err
	}
	if dump_interval < 1 {
		return errors.New("dump_interval is invalid")
	}

	refresh_interval, err = cmd.Flags().GetUint32("refresh_interval")
	if err != nil {
		return err
	}
	if refresh_interval < 1 {
		return errors.New("refresh_interval is invalid")
	}

	return nil
}

func readGeneralClientArgs(cmd *cobra.Command) error {

	if err := readRootArgs(cmd); err != nil {
		return err
	}

	var err error
	report_interval, err = cmd.Flags().GetUint32("report_interval")
	if err != nil {
		return err
	}
	if report_interval < 1 {
		return errors.New("report_interval is invalid")
	}

	nat66Gateway, err = cmd.Flags().GetBool("enable_aux_nat66_stat")
	if err != nil {
		return err
	}



	nat66Persistent, err = cmd.Flags().GetBool("enable_nat66_persistent")
	if err != nil {
		return err
	}

	influxdbUrl, err = cmd.Flags().GetString("influxdb_url")
	if err != nil {
		return err
	}
	if nat66Persistent && influxdbUrl == "" {
		return errors.New("influxdb_url is invalid")
	}

	influxdbOrgBucket, err = cmd.Flags().GetString("influxdb_org_bucket")
	if err != nil {
		return err
	}
	if nat66Persistent && influxdbOrgBucket == "" {
		return errors.New("influxdb_org_bucket is invalid")
	}

	influxdbToken, err = cmd.Flags().GetString("influxdb_token")
	if err != nil {
		return err
	}
	if nat66Persistent && influxdbToken == "" {
		return errors.New("influxdb_token is invalid")
	}

	return nil
}


func main() {

	rootCmd.AddCommand(generalClientCmd, serverCmd)

	if err := rootCmd.Execute(); err != nil {
		//util.MaoLog(util.ERROR, fmt.Sprintf("Fail to execute rootCmd: %s", err))
		os.Exit(-1)
	}
}