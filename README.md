# dice

```
go install github.com/coreyog/dice@latest
```

```
/> dice d6 8D20 2d4+2d6 6d%+6d10 6dF 2d6+10 1d6-1 8-4
invalid dice entry: 2d4_2d6
 d6(  2 )                                                                                                                               =   2
d20( 20 ) + d20( 17 ) + d20( 14 ) + d20( 11 ) + d20(  8 ) + d20(  6 ) + d20( 5 ) + d20( 3 )                                             =  84
 d6(  5 ) +  d6(  2 ) +  d4(  4 ) +  d4(  1 )                                                                                           =  12
 d%( 90 ) +  d%( 70 ) +  d%( 60 ) +  d%( 30 ) +  d%( 20 ) +  d%( 00 ) + d10( 9 ) + d10( 8 ) + d10( 6 ) + d10( 5 ) + d10( 2 ) + d10( 1 ) = 301
 d6(  4 ) +  d6(  2 ) +        10                                                                                                       =  16
 d6(  5 ) -         1                                                                                                                   =   4
 dF(  1 ) +  dF(  1 ) +  dF(  0 ) +  dF(  0 ) +  dF( -1 ) +  dF( -1 )                                                                   =   0
        4                                                                                                                               =   4
                                                                                                                                        = 423
```

Prints dice throws in same order as args. Prints dice in throw by highest face count first, then by value rolled.

    d6

Rolls 1 die with 6 faces. A normal board game die roll with the outcomes of 1 through 6 inclusive.

    8D20

Rolls 8 distinct die with 20 sides. The output is printed in decreasing order and the sum is added onto the end. Note the capital D means nothing.

    2d4+2d6

Rolls 2 die with 4 sides and 2 die with 6 sides. It'll print the outputs of the d6s first, followed by the d4s, followed by a total.

    6d%+6d10

Also supports the percentile dice as rolling outcomes 00, 10, 20, 30, 40, 50, 60, 70, 80, 90.

    6dF

Supports Fudge/Fate Dice. A d6 is rolled and mapped to [-1, 1] (1-2 => -1, 3-4 => 0, 5-6 => 1).

    2d6+10

Constants can be added to a group.

    1d6-1

Constants can be negative too.

    8 - 4

Constants can be added or subtracted and are consolidated. Dice can only be added.

Finally a total of all throws: 423.