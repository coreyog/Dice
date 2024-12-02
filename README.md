# dice

```
go install github.com/coreyog/dice@latest
```

```
/> dice d6 3D20 2d4_2d6 d%+d10 6dF 2d6+10 8-4
invalid dice entry: 2d4_2d6
 d6(  2 )                                                        =   2
d20( 17 ) + d20( 12 ) + d20( 2 )                                 =  31
 d%( 70 ) + d10(  7 )                                            =  77
 dF(  1 ) +  dF(  1 ) +  dF( 1 ) + dF( 0 ) + dF( -1 ) + dF( -1 ) =   1
 d6(  1 ) +  d6(  1 ) +       10                                 =  12
        4                                                        =   4
                                                                 = 127
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

    6dF

Supports Fudge/Fate Dice. A d6 is rolled and mapped to [-1, 1] (1-2 => -1, 3-4 => 0, 5-6 => 1).

    2d6+10

Constants can be added to a group.

    8 - 4

Constants can be added or subtracted and are consolidated. Dice can only be added.

Finally a total of all throws: 127.