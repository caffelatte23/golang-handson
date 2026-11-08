package errsqrt

import (
	"math"
	"testing"
)

func TestSqrt(t *testing.T) {
	if _, ok := any(ErrNegativeSqrt(0)).(error); !ok {
		t.Fatal("ErrNegativeSqrt does not implement error (Error() string が必要です)")
	}
	v, err := Sqrt(2)
	if err != nil || math.Abs(v-math.Sqrt2) > 1e-9 {
		t.Errorf("Sqrt(2) = %v, %v", v, err)
	}

	_, err = Sqrt(-2)
	if err == nil {
		t.Fatal("Sqrt(-2) should return error")
	}
	if got, want := err.Error(), "cannot Sqrt negative number: -2"; got != want {
		t.Errorf("err.Error() = %q, want %q", got, want)
	}
	// 返されたエラーの具体的な型が ErrNegativeSqrt であること
	if e, ok := any(err).(ErrNegativeSqrt); !ok || float64(e) != -2 {
		t.Errorf("err should be ErrNegativeSqrt(-2), got %T(%v)", err, err)
	}
}
