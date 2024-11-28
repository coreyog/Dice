# dice

```
go install github.com/coreyog/dice
```

```
/> dice d6 3D20 2d4+2d6 d%+d10
  d6(  6 )
 d20( 12 ) + d20( 10 ) + d20( 3 )           = 25
  d6(  6 ) +  d6(  5 ) +  d4( 2 ) + d4( 2 ) = 15
  d%( 00 ) + d10(  8 )                      =  8
Total:  48
```

Prints dice throws in same order as args. Prints dice in throw by highest face count first, then by value rolled.

    d6

Rolls 1 die with 6 faces. A normal board game die roll with the outcomes of 1 through 6 inclusive.

		3D20

Rolls 3 distinct die with 20 sides. The output is printed in decreasing order and the sum is added onto the end. Note the capital D means nothing.

		2d4+2d6

Rolls 2 die with 4 sides and 2 die with 6 sides. It'll print the outputs of the d6s first, followed by the d4s, followed by a total.

		d%+d10

Also supports the percentile dice as rolling outcomes 00, 10, 20, 30, 40, 50, 60, 70, 80, 90.

Finally a total of all throws: 48.