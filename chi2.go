package classifier

import "math"

func invChi2(x2 float64, df int) float64 {
        m := x2 / 2.0
        sum := math.Exp(-m)
        term := sum
        for i := 1; i < df/2; i++ {
                term *= m / float64(i)
                sum += term
        }
        if sum > 1.0 {
                sum = 1.0
        }
        return sum
}
