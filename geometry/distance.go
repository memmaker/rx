package geometry

// code of this file is a modified version of code from
// https://github.com/anaseto/gruid, which has the following license:
//
// Copyright (c) 2020 Yon <anaseto@bardinflor.perso.aquilenet.fr>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

func DistanceSquared(p, q Point) int {
	p = p.Sub(q)
	return p.X*p.X + p.Y*p.Y
}

// DistanceManhattan computes the taxicab norm (1-norm). See:
//
//	https://en.wikipedia.org/wiki/Taxicab_geometry
//
// It can often be used as A* distance heuristic when 4-way movement is used.
func DistanceManhattan(p, q Point) int {
	p = p.Sub(q)
	return Abs(p.X) + Abs(p.Y)
}

// DistanceChebyshev computes the maximum norm (infinity-norm). See:
//
//	https://en.wikipedia.org/wiki/Chebyshev_distance
//
// It can often be used as A* distance heuristic when 8-way movement is used.
func DistanceChebyshev(p, q Point) int {
	p = p.Sub(q)
	return max(Abs(p.X), Abs(p.Y))
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}
