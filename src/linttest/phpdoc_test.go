package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestBadPhpdocTypes(t *testing.T) {
	// If there is an incorrect phpdoc annotation,
	// don't use it as a type info.
	//
	// Before the fix, NoVerify inferred \$a and \$b to be
	// types for corresponding params, which is incorrect.

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @param $a
 * @param $b
 * @return int
 */
function fav_func($a, $b) {
  if ($a[0] != $b[0]) {
    return ($a[0] > $b[0]) ? -1 : 1;
  }
  return ($a[1] < $b[1]) ? -1 : 1;
}
`)
	test.Expect = []string{
		`PHPDoc is incorrect: malformed @param $a tag (maybe type is missing?)`,
		`PHPDoc is incorrect: malformed @param $b tag (maybe type is missing?)`,
	}
	test.RunAndMatch()
}

func TestPHPDocPresence(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	trait TheTrait {
		public function traitPub() {}
	}
	class TheClass {
		/**
		 * This function is a good example.
		 * It's public and has a phpdoc comment.
		 */
		public function documentedPub() {}

		// Not OK.
		public function pub() {}

		// OK.
		private function priv() {}

		// OK.
		protected function prot() {}
	}`)
	test.Expect = []string{
		`Missing PHPDoc for "pub" public method`,
		`Missing PHPDoc for "traitPub" public method`,
	}
	test.RunAndMatch()
}

func TestPHPDocSyntax(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	/**
	 * @param $x int the x param
	 * @param - $y the y param
	 * @param $z - the z param
	 * @param $a
	 * @param int
	 */
	function f($x, $y, $z, $a, $_) {
		$_ = $x;
		$_ = $y;
		$_ = $z;
	}`)
	test.Expect = []string{
		`non-canonical order of variable and type on line 2`,
		`expected a type, found '-'; if you want to express 'any' type, use 'mixed' on line 3`,
		`non-canonical order of variable and type on line 4`,
		`expected a type, found '-'; if you want to express 'any' type, use 'mixed' on line 4`,
		`malformed @param $a tag (maybe type is missing?) on line 5`,
		`malformed @param tag (maybe var is missing?) on line 6`,
	}
	test.RunAndMatch()
}

func TestPHPDocType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	/**
	 * @param [][]string $x1
	 * @param double $x2
	 * @param real $x3
	 * @param integer $x4
	 * @param boolean $x5
	 * @return []int
	 */
	function f($x1, $x2, $x3, $x4, $x5) {
		$_ = [$x1, $x2, $x3, $x4, $x5];
		return [1];
	}`)
	test.Expect = []string{
		`[]int type syntax: use [] after the type, e.g. T[]`,
		`[][]string type syntax: use [] after the type, e.g. T[]`,
		`use float type instead of double`,
		`use float type instead of real`,
		`use int type instead of integer`,
		`use bool type instead of boolean`,
	}
	test.RunAndMatch()
}
