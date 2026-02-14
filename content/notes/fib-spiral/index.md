+++
date = '2026-02-14T10:54:29-05:00'
draft = false
title = 'generating a fib spiral'
summary = 'Neat trick to tile a fib spiral using complex numbers and representing the fib number as a vector.'
tags = ['programming', 'graphics']
toc_sidebar = true
custom_toc = """contents
[tiling](#tiling)
[spirals](#spirals)
files
[fib_spiral_cli.js](fib_spiral_cli.js)
[fib-12.svg](fib-12.svg)
[fib-12_squares.svg](fib-12_squares.svg)"""
+++

# tiling
A neat way to generate a fib spiral is to think of it in terms of vectors. Take the fib sequence
```
0,1,1,2,3,5,8,13,21...
```
and represent each fib num, `Fn`, as the vector `<Fn, Fn>`. In this particular case, it will help us to think of this as a complex vector. So the vector is really `Fn + Fn*i`: `Fn` units in the positive real line, and `Fn` units in the positive imaginary line.

Now we we can multiply it by another complex number, say `0 - i` or `0 + i`. In these cases we are effectively rotating this vector by 90° either clockwise or counterclockwise, respectively. The image below demonstrates this when multiplying `Fn + Fn*i` by `0 - i`. It only rotates the vector, 90° clockwise in this case.

{{< figure class="note-center" src="rotation.svg" >}}

We can also rotate by 180° with `-1 + 0i`. Lastly, no rotation is done by `1 + 0i`. Below is a table summarizing this:

| rot       | Effect                      |
| --------- | --------------------------- |
| `1 + 0i`  | no rotation                 |
| `0 - i`   | rotate 90° clockwise        |
| `-1 + 0i` | rotate 180°                 |
| `0 + i`   | rotate 90° counterclockwise |

Using this knowlege we can now take each `Fn + Fn*i` and cycle through the above table rotations. This will allow us to tile the fib points in a spiral pattern, emanating outward. Using the SVG path and the arc subcommand, we can build the spiral connecting two fib points like 

```
<path class="arc" d="M ${x0} ${y0} A ${Fn} ${Fn} 0 0 0 ${x1} ${y1}" />
```

# spirals

Below is the result of this tiling pattern using arcs. We get a satisying spiral from the fib vectors. In the second image, the fib squares each have dimension `Fn` x `Fn` and illustrate how the arcs are formed, each inside a fib square, connected by the 2 diagonal points of the square.

{{< figure class="note-center" src="fib-12.svg" caption="fib spiral, 12 segments" >}}

{{< figure class="note-center" src="fib-12_squares.svg" caption="fib spiral, 12 segments, squares" >}}

I used [fib_spiral_cli.js](fib_spiral_cli.js) to generate the two spirals above.

[top](#)