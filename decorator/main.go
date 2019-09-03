package main

import (
	"fmt"
)

type Config struct {
	Name    string
	Address string
	Age     int
}

type OptionFunc func(*Config)

func WithName(name string) OptionFunc {
	return OptionFunc(func(c *Config) {
		c.Name = name
	})
}

func WithAddress(address string) OptionFunc {
	return OptionFunc(func(c *Config) {
		c.Address = address
	})
}

func WithAge(age int) OptionFunc {
	return OptionFunc(func(c *Config) {
		c.Age = age
	})
}

func ApplyConfig(c *Config, opts ...OptionFunc) {
	for _, opt := range opts {
		opt(c)
	}
}

func main() {
	c := &Config{}

	ApplyConfig(c, WithAge(25), WithName("YaoZengzeng"), WithAddress("Hangzhou"))

	fmt.Printf("Name: %v, Address: %v, Age: %v\n", c.Name, c.Address, c.Age)
}
