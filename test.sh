#!/bin/bash
assert() {
  expected="$1"
  input="$2"

  ./chibicc "$input" > tmp.s
  cc -o tmp tmp.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assertErr() {
  expected="$1"
  input="$2"

  ./chibicc "$input" 2> tmp.err
  
  actual=$(cat tmp.err)
  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo $actual
    echo $expected
    diff <(echo "$actual") <(echo "$expected")
    exit 1
  fi
}

assert 0 0
assert 42 42
assert 41 " 12 + 34 - 5 "
assertErr "1x1
 ^ 
cannot tokenize x" "1x1"

echo OK