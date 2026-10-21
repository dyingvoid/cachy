package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	infrastructure_resp3 "github.com/dyingvoid/cachy/internal/infrastructure/resp3"
	transport_tcp_client "github.com/dyingvoid/cachy/internal/transport/tcp/client"
)

func main() {
	host := flag.String("h", "localhost", "server host")
	port := flag.Int("p", 8081, "server port")
	flag.Parse()

	if err := run(*host, *port); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(host string, port int) error {
	cfg, err := transport_tcp_client.NewConfig(host, port)
	if err != nil {
		return err
	}

	client, err := transport_tcp_client.NewClient(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Printf("connected to %s:%d\n", host, port)
	fmt.Println("commands: PING | INFO | GET key | SET key value [EX|PX ttl] | LPUSH key value... | RPUSH key value... | LPOP key count | LRANGE key start stop | LLEN key | REPLCONF listening-port|capa value | PSYNC id offset")
	fmt.Println("empty line to quit")

	parser := features_commands.NewProtocolParser()
	serializer := &infrastructure_resp3.Serializer{}
	deserializer := infrastructure_resp3.NewDeserializer(client)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			fmt.Println()
			return scanner.Err()
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil
		}

		cmd, err := parser.Parse(line)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		if _, err := client.Write(serializer.SerializeCommand(cmd)); err != nil {
			return fmt.Errorf("send command: %w", err)
		}

		value, err := deserializer.Read()
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		fmt.Println(formatValue(value))
	}
}

func formatValue(value infrastructure_resp3.Value) string {
	switch value.Kind {
	case infrastructure_resp3.KindSimpleString, infrastructure_resp3.KindBulkString:
		return value.String
	case infrastructure_resp3.KindSimpleError, infrastructure_resp3.KindBulkError:
		return "(error) " + value.String
	case infrastructure_resp3.KindInteger:
		return strconv.FormatInt(value.Integer, 10)
	case infrastructure_resp3.KindBoolean:
		return strconv.FormatBool(value.Boolean)
	case infrastructure_resp3.KindDouble:
		return strconv.FormatFloat(value.Double, 'g', -1, 64)
	case infrastructure_resp3.KindBigNumber:
		return value.BigNumber.String()
	case infrastructure_resp3.KindNull:
		return "(nil)"
	case infrastructure_resp3.KindVerbatimString:
		return value.Encoding + ":" + value.String
	case infrastructure_resp3.KindArray, infrastructure_resp3.KindSet, infrastructure_resp3.KindPush:
		elements := make([]string, 0, len(value.Elements))
		for _, element := range value.Elements {
			elements = append(elements, formatValue(element))
		}
		return "[" + strings.Join(elements, ", ") + "]"
	case infrastructure_resp3.KindMap:
		pairs := make([]string, 0, len(value.Pairs))
		for _, pair := range value.Pairs {
			pairs = append(pairs, formatValue(pair.Key)+" => "+formatValue(pair.Value))
		}
		return "{" + strings.Join(pairs, ", ") + "}"
	default:
		return fmt.Sprintf("%#v", value)
	}
}
