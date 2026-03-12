package fmt

import "testing"

func TestLoadBytesInt64(t *testing.T) {
	c := GetConv()
	defer c.PutConv()
	c.LoadBytes([]byte("42"))
	v, err := c.Int64()
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestLoadBytesFloat64(t *testing.T) {
	c := GetConv()
	defer c.PutConv()
	c.LoadBytes([]byte("9.5"))
	v, err := c.Float64()
	if err != nil {
		t.Fatal(err)
	}
	if v != 9.5 {
		t.Errorf("expected 9.5, got %f", v)
	}
}

func TestLoadBytesNegative(t *testing.T) {
	c := GetConv()
	defer c.PutConv()
	c.LoadBytes([]byte("-100"))
	v, _ := c.Int64()
	if v != -100 {
		t.Errorf("expected -100, got %d", v)
	}
}

func TestLoadBytesScientific(t *testing.T) {
	c := GetConv()
	defer c.PutConv()
	c.LoadBytes([]byte("1.5e2"))
	v, _ := c.Float64()
	if v != 150 {
		t.Errorf("expected 150, got %f", v)
	}
}

func BenchmarkLoadBytesInt64(b *testing.B) {
	data := []byte("12345")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := GetConv()
		c.LoadBytes(data)
		c.Int64()
		c.PutConv()
	}
}

func BenchmarkLoadBytesFloat64(b *testing.B) {
	data := []byte("3.14159")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := GetConv()
		c.LoadBytes(data)
		c.Float64()
		c.PutConv()
	}
}
