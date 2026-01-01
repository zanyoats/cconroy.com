+++
date = '2026-01-01T10:22:29-05:00'
draft = false
title = 'chapter notes for the little typer'
summary = 'These notes serve as an addendum to each chapter. I wish I had these notes alongside each chapter while working through the book.'
+++

{{< figure src="9780262536431.avif" class="note-right" link="https://mitpress.mit.edu/9780262536431/the-little-typer/" caption="[the little typer](https://mitpress.mit.edu/9780262536431/the-little-typer/), by MIT Press" >}}

That was quite the odyssey into the arcane realm of dependent types. At the outset, I was not interested in the logic itself, but in the power it can confer on programming languages. The types constructed are much more expressive than we see in traditional statically typed languages, à la Java, C/C++, OCAML, Scala, and Haskell. In dependently typed languages, construction can be interpreted as a form of proof--that is, finding a program P with type T is proof of the logical statement that T encodes. Claiming T is all fine and good, but proving it requires finding a program that the compiler accepts.

That might be the paradigm shift this type of language makes. But as good as this sounds, they might not be practical for most programming needs. They probably don't scale well when developing much larger programs like OSes, games, browsers, etc. Additionally, on the other end of the spectrum, they are not worth the time for simple scripts. Nevertheless, they are worth knowing about. If you want to proceed further down the rabbit hole, graduating to more advanced dependently-typed languages used for making real programs, then roughly, there seems to be two paths: (1) developing runnable programs, and proving properties about them, or (2) developing theorems just to prove them; essentially doing mathematics with machine-checked proofs. Idris has positioned itself for the former with Roq/Lean/Agda purpose-built for the latter. For me, looking into Idris would be a next step--at least just to see what real programs are being built using this newfound knowledge.

As a reference, the [little_typer_dialog.pie](little_typer_dialog.pie) is the code (most of it) the book uses to teach you this new language. Below are the notes by chapter(s):

1. [Chapters 1-2: Getting Definitions Down](#chapters-1-2-getting-definitions-down)
1. [Chapter 3: What is a Proof, What is a Value](#chapter-3-what-is-a-proof-what-is-a-value)
1. [Chapter 4: ->'s Stronger Brother Pi](#chapter-4--s-stronger-brother-pi)
1. [Chapters 5-6: Lists and Vectors](#chapters-5-6-lists-and-vectors)
1. [Chapter 7: ind-Nat is a Dependent rec-Nat](#chapter-7-ind-nat-is-a-dependent-rec-nat)
1. [Chapters 8-9: Proving Properties About Programs](#chapters-8-9-proving-properties-about-programs)
1. [Chapters 10-11: Induction on Lists and Vecs](#chapters-10-11-induction-on-lists-and-vecs)
1. [Chapters 12-13: Even and Odd](#chapters-12-13-even-and-odd)
1. [Chapter 14: The Absurd Type](#chapter-14-the-absurd-type)
1. [Chapters 15-16: CL = IL + DNE](#chapters-15-16-cl--il--dne)

### Chapters 1-2: Getting Definitions Down

In the preface, the text mentions what is meant by **Laws** and **Commandments**. It's worth spending some time to digest these terms before proceeding. You will need to do as such in the first few chapters since much of this text relies on understanding the precise meanings of its definitions. But don't sway into thinking this is an intractable computer science monograph on type theory--it strikes a good balance between not being too academic and still allowing you to swim in the deep end of the pool. Very much living up to its playful artwork and cover.

With this framing in mind, now on to the definitions! Laws determine what expressions mean, while Commandments determine when two expressions mean the same thing. The first law you encounter (The Law of Tick Marks) defines which expressions are atoms. `'pear` is an atom in accord with this law but the expression `'42` is not. Next you will see the Commandment of Tick Marks indicating when two atoms are the same. You will want to make sure you understand the laws before looking at their respective commandments.

Another concept worth mulling on is what is meant by a **judgment**. Looking at the first four templates provides some insight. Filling in all the blanks in a template produces a judgment. Whether or not that judgment is believable is a different concern.

1. ______ is a ______.
1. ______ is type.
1. ______ is the same ______ as ______.
1. ______ and ______ are the same type.

I should have mentioned earlier that the language used in the book is called Pie. These judgments describe the rules the typechecker must follow. They determine whether computations are well-typed and therefore meaningful. Each judgment relies on one or more presuppositions. For example, the form `______ is a ______.` requires:

* the first blank to be an expression
* the second blank to be an expression that is a type.

`'pear is a Atom.` is an example of a judgment that is believable but `'pear is a 'fruit.` or `'pear is a Pair(Atom, Atom).` are not. In both cases we don't believe them: `'fruit` is not a type and an Atom is not the same as a Pair.

In Pie everything is a pure expression except there are two special forms: `claim` and `define`. Both accept a name and an expression as arguments. `claim` asserts the type by name while `define` associates a name with the expression. Every `define` needs a corresponding `claim`. The text then goes on to describe the different kinds of expressions: **values**, **normal forms**, and **neutral**. Values and normal forms are straightforward concepts that don't need much elaboration, but an example will suffice:

```scheme
; a value, since it has the add1 constructor at its top
(add1 (+ zero (add1 zero)))

; also the same value, but in its reduced normal form
(add1 (add1 zero))
```

Neutral expressions are more subtle. They are not values, and cannot be reduced any further due to an unknown variable. They are kind of stuck in limbo or purgatory until context is provided.

```scheme
; context is:
;   y is the Atom 'fruit
;   x is a (Pair Atom Atom)
;   z is a Nat

; neutral expressions
(car x)
(car (cons x y))
z 

; these are not, since they are values because
; the constructor is at the top
(add1 (+ z z))
(cons (car x) (cdr x))
```

One minor nuisance in these first chapters is that many expressions are well-formed but not directly runnable without type annotations. The following example won't work:

```scheme
; will not compile :(
(car
    (cons
        (cons 'aubergine 'courgette)
        'tomato))
```

It will complain about not being able to determine the type. You can annotate the expression with `the`:

```scheme
(car
    (the (Pair (Pair Atom Atom) Atom)
        (cons (cons 'aubergine 'courgette)
            'tomato)))
```

Types are themselves expressions, and their type is `U` (universe). The type for `Nat`, `Atom`, etc. is `U`. Type expressions being treated like value expressions is what gives types first-class privileges in Pie. You can use `U` with `the` like below:

```scheme
(Pair
    (car (the (Pair U Atom)
        (cons Atom 'olive)))
    (cdr (the (Pair Atom U)
        (cons 'oil Atom))))
```

Lastly, some of the examples use the `+` operator. This will be defined in future chapters; for now, you can download [add.pie](add.pie) to run those examples and follow along until addition is introduced properly.

[top](#)

### Chapter 3: What is a Proof, What is a Value

There is a subtle connection between computation and traditional mathematical logic:

* programs are proofs,
* types are statements (or propositions), and
* to find a proof is equivalent to writing a program that the compiler accepts.

In Pie all programs are expressions that are either values or almost values (neutral). A value can be viewed as a proof. Proof of what? Proof of whatever its type is. What about the case when an expression has no value? That means there is no proof, or whatever the statement is encoded in its type is false. An example will solidify this:

```scheme
; You haven't seen `=` but it is a type constructor (like Pair)
; where its arguments are equal w.r.t. some type.
;
; Pie will allow this claim (below) but coming up with a program
; for it will never work, but let's relax that for now.

(claim bogus
    (-> Nat
        (= Nat zero (add1 zero))))

(define bogus
    (lambda (n)
        (bogus n)))
```

Let's assume Pie allowed recursion. Then such a definition would be completely reasonable. In the body of `bogus` I need to return the type `(= Nat zero (add1 zero))`. Well, why not return `(bogus n)` which conveniently provides this exact type.

* _But_ this results in an infinite recursion and never produces a value
* _But_ according to Pie, the definition of bogus is reasonable proving the claim that if I provide any Nat, then 0 must equal 1

Pie resolves this fatal flaw by disallowing recursion in this manner. With this omission, programs now produce values, and can be stashed in types and the Pie compiler can be assured that computing types would never result in an endless loop.

[top](#)

### Chapter 4: ->'s Stronger Brother Pi

Our friend `->` has already been used to type lambdas. But now we get introduced to `Pi`, which gives us what other languages call generics. With it, we can now define an eliminator for the generic `(Pair A D)` where `A` and `D` are any types in `U`:

```scheme
(claim elim-Pair
  (Pi ((A U) (D U) (X U))
    (-> (Pair A D)
        (-> A D
            X)
        X)))

(define elim-Pair
  (lambda (A D X p f)
    (f (car p) (cdr p))))
```

`Pi` is like `->` but can also take on arguments that get substituted in `Pi` bodies. `(elim-Pair Nat Atom Nat)` computes the type:

`(-> (Pair Nat Atom) (-> Nat Atom Nat) Nat)`

which describes some lambda expression. Something `->` was not able to do on its own.

[top](#)

### Chapters 5-6: Lists and Vectors

If we wanted to create list of things and we only know about pairs one way to encode this (tediously) is:

```scheme
; Avoiding empty lists, `base-num-type` is the type for
; lists of one element. 
(claim base-num-type U)
(define base-num-type (Pair Nat Atom))

; Make our list type `Num` one element larger by wrapping it
; inside a new pair type with a `Nat`.
(claim cons-num-type
  (Pi ((Num U))
    U))
(define cons-num-type
  (lambda (Num)
    (Pair Nat Num)))

; Constructor for creating one element lists
(claim base-num
  (-> Nat
      base-num-type))
(define base-num
  (lambda (n)
    (cons n 'nil)))

; Constructor for creating lists one element larger
(claim cons-num
  (Pi ((num-type U)
       (nums num-type))
    (-> Nat
        (cons-num-type num-type))))
(define cons-num
  (lambda (num-type nums num)
    (cons num nums)))
```

Here is a session log creating a one element list and consing to it twice:

```scheme
> (base-num 69)
(the (Pair Nat Atom)
  (cons 69 'nil))

> (cons-num base-num-type
    (base-num 69)
    42)
(the (Pair Nat
       (Pair Nat Atom))
  (cons 42
    (cons 69 'nil)))

> (cons-num (cons-num-type base-num-type)
    (cons-num base-num-type
        (base-num 69)
        42)
    0)
(the (Pair Nat
       (Pair Nat
         (Pair Nat Atom)))
  (cons 0
    (cons 42
      (cons 69 'nil))))
```

As you can see this is not workable. In these chapters we are introduced to two new list types: lists and vectors. Lists are cons structures with type `(List E)` for some `E` in `U` along with the constructors `::` and `nil`. It has one eliminator named `rec-List`. Vectors are lists, but their corresponding length is part of the type--making it the first **dependent type** introduced. Its type looks like `(Vec E k)` for some `E` in `U` and `k` in `Nat`. It has constructors `vec::` and `vecnil`. The only eliminators presented are `head` and `tail`. At the moment we have no eliminator as powerful as the `rec-List` analog.

The most salient concept in these chapters is that of a **total function**. Let's claim I have a function `bogus-first` that takes the head of a vector. As described here `bogus-first` cannot be implemented in Pie because it's not total. There is no `(bogus-first vecnil)` program/value so Pie will reject the following claim:

```scheme
(claim bogus-first
  (Pi ((E U) (l Nat))
    (-> (Vec E l)
        E)))

; will not compile since `es` type does not have `add1` at top
;(define bogus-first
;  (lambda (E l es)
;    (head es)))
```

One simple approach to resolve this is to make `first` return `(Option E)` instead of `E` which would now make it total. But Pie mentions a more idiomatic approach in the principle _Use a More Specific Type_. Make `first` take a vector with at least one element in it. That would look like:

```scheme
(claim first
  (Pi ((E U)
       (l Nat))
    (-> (Vec E (add1 l)) ; <-- vectors of length l+1 as input
        E)))
(define first
  (lambda (E l es)
    (head es)))
```

Now `first` is total, since it works on all vectors of at least one element.

[top](#)

### Chapter 7: ind-Nat is a Dependent rec-Nat

We meet `ind-Nat` for when `rec-Nat` is not up to the task. The trivial example to motivate us quickly is: building a vec of atoms with a particular length in mind. `rec-Nat` is not capable of implementing this simple program since it cannot vary the types depending on the target:

```plaintext
> start with an empty vector to get
  vecnil
  with type (Vec Atom 0)

> add one 'a in the step to get
  (vec:: 'a vecnil)
  with type (Vec Atom 1)

> add one 'a in the step to get
  (vec:: 'a (vec:: 'a vecnil))
  with type (Vec Atom 2)

> ... do this n-3 more times ...

> add one 'a in the step to get
  (vec:: 'a ... (vec:: 'a (vec:: 'a vecnil))...)
  with type (Vec Atom n)
```

As you can see the type changes in each step of the recursion, thus we need a dependent type that takes Nats from [0, n]. `ind-Nat` accepts this dependent type as one of its arguments. This argument is called the *motive* and is used in other inductive expressions.

We have come a long way since `which-Nat`. `iter-Nat` was needed when `which-Nat` could not do recursion. `rec-Nat` combined both `iter-Nat` and `which-Nat` for when you might want access to each element in the recursion. Now, it is superseded by `ind-Nat` which can do everything prior, plus more, when dependent types are involved.

[top](#)

### Chapters 8-9: Proving Properties About Programs

The difficulty ramps up a bit in these chapters. The type constructor equals (i.e. `=`) is introduced along with its `same` constructor and a cadre of eliminators: `cong`, `replace`, and `symm`. `same` is a way to construct evidence for an expression, which is what is needed at the base case of induction. During each step of the induction we need eliminators to construct `evidence_k` from given `evidence_k-1`. If we can achieve this we have `evidence_n`, or proof of the statement the type encodes. 

Let's start with `cong` since it's simpler than `replace`. It takes two arguments. The first is `evidence_k-1` which has type `(= X from to)`. If you want to prove `(= X (f from) (f to))` and you can find some `f`, then `evidence_k` will be `(cong evidence_k-1 f)`. Here is the example from the book using `cong`:

```scheme
; base case
(claim base-incr=add1
  (= Nat (incr zero) (add1 zero))) ; (= Nat 1 1)
(define base-incr=add1
  (same 1)) ; same constructor for base case

; the step proves that:
;   for every Nat n
;   if (incr n) equals (add1 n)
;   then (incr (add1 n)) equals (add1 (add1 n))

(claim step-incr=add1
  (Pi ((n-1 Nat))
    (-> (= Nat
          (incr n-1)
          (add1 n-1))
        (= Nat
          (add1 (incr n-1))
          (add1 (add1 n-1))))))

(define step-incr=add1
  (lambda (n-1)
    (lambda (evidence_k-1)      ; given evidence
      (cong                     ; cong eliminator for inductive step
        evidence_k-1 (+ 1)))))  ; evidence_k with f = (+ 1)
```

Understanding these chapters required multiple readings. If `cong` is not clear, take a break and start the chapter again with a fresh mind.

`replace` is a complex beast. At first it is not clear what it does. When to use it is simpler since `cong` can only work on and return equality types. `replace` can work on an equality type but return any type described by its motive. It takes three arguments with the first being the target evidence with type `(= X from to)`. The third argument is the base, which is an expression that is close to the one you want, but you need to do some minor surgery to get it just right. The middle argument is the motive, which is a function that takes `from` and `to` from the target evidence and uses them to construct the types for the base and whole replace expression, respectively.

The example to implement `twice-Vec` should clear up `replace`, and what purpose each argument is for. Below we implement `twice-Vec`, making use of the proof `twice=double` and the function `double-Vec`. Using these parts, `replace` can help us implement `twice-Vec`. `double-Vec` with type `(= Vec E (double l))` is the base that is very close to the type we want: `(= Vec E (twice l))`. We also know twice equals double, but with `symm` we can rearrange the order to double equals twice and pass that as the target to `replace`. `replace` will allow us to return the result of `double-Vec` for `twice-Vec` since it can now transform the type from `(= Vec E (double l))` to `(= Vec E (twice l))` using the target, motive, and base.

```scheme
(define twice-Vec
  (lambda (E l)
    (lambda (es)
      (replace
        (symm (twice=double l)) ; <-- target
        (lambda (k)             ; <-- motive
          (Vec E k))
        (double-Vec E l es))))) ; <-- base
```

[top](#)

### Chapters 10-11: Induction on Lists and Vecs

We are introduced to a new type constructor: `Sigma` -- which encodes the _there exists_ operator from classical logic. As a recap we can now say the following in Pie's type system:

1. Whether something is a `Nat`, `Atom`, `List`, `Vec` or `Pair` (which is `Sigma`)
1. Whether both statements are true (and) ... using `Pair`
1. Either or statement (or) ... using `Either` (introduced in Chapter 13)
1. Equivalence, if and only if ... using `=`
1. If P is true, then Q is true (P implies Q) ... using `->`
1. For all ... using `Pi`
1. There exists ... using `Sigma`

With `Sigma` now in our toolbox, we set out to implement a new helper: `list->vec`. While possible, this results in specious definitions such as vecs that have a length that don't correspond to the input list. We overcame this deficiency by tying the length of the list argument to the vector length as its type:

```scheme
(claim list->vec
  (Pi ((E U)
       (es (List E)))
    (Vec E (length E es)))) ; <-- vec has length of `es` now
```

But this means we need to upgrade from `rec-List` to `ind-List` to complete the implementation. `ind-List` will do induction on each element of the list and the motive argument will allow us to return correct vector type in each step:

```scheme
(define mot-list->vec
  (lambda (E)
    (lambda (es) ; <-- each step has the vec length
      (Vec E (length E es)))))
```

Chapter 10 concludes that even this definition suffers from specious implementations. For example, returning an appropriate-sized vector but the order is reversed, or just some of the elements are swapped. Unfortunately, the type doesn't preclude these definitions. To resolve this, we need not change the type again, but delegate to an external proof to help:

```scheme
; For every list `es`
;   `es` is equal to itself converted to
;   a vec and then back to a list

(claim list->vec->list=
  (Pi ((E U)
       (es (List E)))
    (= (List E)
       es
       (vec->list E
         (length E es)
         (list->vec E es)))))
```

We proceed to make a definition for `list->vec->list=`, which is proof that the list turned into a vec using `list->vec` is indeed a vec with the same elements and order.


[top](#)

### Chapters 12-13: Even and Odd

These chapters don't introduce any new concepts (except `(Either L R)`), but build on our knowledge to develop the dependent types: `(Even n)` and `(Odd n)`. We proceed to prove very simple properties about even and odd numbers:

* adding 1 to an even produces an odd
* adding 1 to an odd produces an even
* every number is either even or odd (using the above proofs)

In theory, you can now employ these types in programs. For example, you could want a function that takes in a number `(Even n)` and doubles it:

```scheme
(claim even-doubler
  (Pi ((n Nat))
    (-> (Even n)
      (Even (double n)))))

(claim twelve-is-even (Even 12))
(define twelve-is-even
  (cons 6 (same 12)))

; it can be called like (even-doubler 6 (Even 6)) and it will return twelve-is-even
; which has the type (Even 12)
```

[top](#)

### Chapter 14: The Absurd Type

If the type `Trivial` has one value, then what type has zero values? `Absurd` does; a program with a type that has `Absurd` in it cannot be defined. Strangely, it does have an eliminator (no constructor) in `ind-Absurd`, but this is to satisfy the type checker in expressions. A clever demonstration of this is in defining `vec-ref`. The types involved here are rather complex. But we want a vector of length 0 to have no 2 index position in `(vec-ref Atom 0 2 vecnil)`, and `Absurd` will work here to let the compiler flag this as an error. Getting the head position of a vector of length 2 should be no problem and return `Atom`: `(vec-ref Atom 2 0 vec-of-len-2)`.

To define `vec-ref` we will need to create a type of numbers that index the vector. The numbers will have a notion of the position in the vector they point at, as well as the length of that vector too. Each layer down in the index number will work with a vector of length one less. The length in the index number will also match the same length of the vector we are trying to index. Such a type of `vec-ref` could look like:

```scheme
(claim vec-ref
  (Pi ((E U)
       (l Nat))
    (-> (Fin l) (Vec E l)
        E)))
```

`(Fin l)` is the type of this index number for some length `l`, of which is the same length of the vector `(Vec E l)`. In coordination we can convince the type checker that we can safely index this vector.

It is now obvious that the target is `l` and we should implement this using `ind-Nat`. The base case will never be used since for it to typecheck the number `(Fin l)` must match the length of the vector. So this is safe by construction. However, `ind-Nat` still requires a base case and an expression using `ind-Absurd` returning `E` will appease the type checker:

```scheme
(claim base-vec-ref
  (Pi ((E U))
    (-> (Fin zero) (Vec E zero)
        E)))
(define base-vec-ref
  (lambda (E no-value-ever es)
    (ind-Absurd no-value-ever
      E)))
```

The inductive step on `ind-Nat` is about peeling off layers of the index number `l`. But I haven't mentioned how this index number or its type look like. This is the most clever part: `(Fin l)` is a dependent type that builds:

* Absurd when l=0
* (Either Absurd Trivial) when l=1
* (Either (Either Absurd Trivial) Trivial) when l=2
* and so on ...

So if we have a vector of length 3 then `(Fin 3)` will have 3 Eithers nested in its type, which in this case represents a vector with 3 valid index positions. The index positions have constructors: `fzero` and `fadd1`. `fzero` represents the index number term for the head position, and `fadd1` gives you subsequent index terms, pointing somewhere in the tail position.

Finally, the step uses this representation and `ind-Either` to decide on whether to call `head` or `tail` on the vector. With the base case and step we have all the parts to assemble the definition for `vec-ref`:

```scheme
(define vec-ref
  (lambda (E l)
    (ind-Nat l
      (lambda (k)
        (-> (Fin k) (Vec E k)
            E))
      (base-vec-ref E)
      (step-vec-ref E))))
```

[top](#)

### Chapters 15-16: CL = IL + DNE

Expressing _refutation of P_ or _not P_ was our last remaining logical analog. We can now express that `(+ 2 2)` does not equal `5` as a type:

```scheme
(claim two+two!=five
  (-> (= Nat (+ 2 2) 5)
      Absurd))
```

Constructing this proof (program) is possible but requires some tricks demonstrated in this chapter. But first let's define it using the `use-Nat=` helper:


```scheme
(define two+two!=five
  (lambda (two+two=five)
    (use-Nat= 0 1
      (use-Nat= 1 2
        (use-Nat= 2 3
          (use-Nat= 3 4
            (use-Nat= 4 5 two+two=five)))))))
```

Remember that `lambda` constructs logical _if_. Then it is saying: if we are somehow given evidence that `two+two=five` -- which is a program proving `(+ 2 2)` equals `5` -- then we can use this evidence with another program that produces `Absurd`. It will type check, but you can never call `(use-Nat= 4 5 two+two=five)` since this requires somehow obtaining `two+two=five`. Since the definition type checks we are stating a refutation of: `(+ 2 2)` equals `5`.

`use-Nat=` will take two Nats and evidence they are equal, and return evidence of a number being equal that is one smaller. In the definition of `two+two!=five` we use this fact to be able to call `use-Nat=` with arguments `0` `1` which compute the type `Absurd` from `=consequence`. Again, it is important to realize this can never happen since we cannot construct evidence for `two+two=five`. But _if_ we could then:

* calling `(use-Nat= 4 5 four=five)` provides evidence `three=four`
* calling `(use-Nat= 3 4 three=four)` provides evidence `two=three`
* calling `(use-Nat= 2 3 two=three)` provides evidence `one=two`
* calling `(use-Nat= 1 2 one=two)` provides evidence `zero=one`
* calling `(use-Nat= 0 1 zero=one)` provides evidence `Absurd`!

Below are the definitions of `=consequence` and `use-Nat=` used in `two+two!=five`:

```scheme
(claim =consequence
  (-> Nat Nat
      U))
(define =consequence
  (lambda (n j)
    (which-Nat n
      (which-Nat j
        Trivial
        (lambda (j-1)
          Absurd))
      (lambda (n-1)
        (which-Nat j
          Absurd
          (lambda (j-1)
            (= Nat n-1 j-1)))))))

(claim =consequence-same
  (Pi ((n Nat))
    (=consequence n n)))
(define =consequence-same
  (lambda (n)
    (ind-Nat n
      (lambda (k)
        (=consequence k k))
      sole
      (lambda (n-1 =consequence_n-1)
        (same n-1)))))

(claim use-Nat=
  (Pi ((n Nat)
       (j Nat))
    (-> (= Nat n j)
        (=consequence n j))))
(define use-Nat=
  (lambda (n j)
    (lambda (n=j)
      (replace n=j
        (lambda (k)
          (=consequence n k))
        (=consequence-same n)))))
```

Now what is the meaning of _CL = IL + DNE_? Expanded out it stands for: Classical logic = Intuitionistic Logic + Double Negation Elimination. If you read this chapter it may have been confusing -- it was for me -- when we proved that `pem-not-false` after realizing `pem` probably has no definition. The type system in Pie is based on constructive logic (i.e. Intuitionistic Logic) which deviates from classical logic on the question of whether PEM is valid. In constructive logic, (P ∨ ¬P) requires constructing a proof of either P or ¬P. Classical logic accepts ¬¬P → P, or if something isn't false, then it is true. A constructivist would reject this judgment. They would also say proving ¬¬P is weaker than proving P. Proving ¬¬P means that _P is false_ leads to a contradiction, but that does not mean P is proved. It just means we have ruled out P being false. However, we still need to prove P by construction to "exclude the middle" so to speak.

The form ¬¬P is why we can define `pem-not-false` below. It is saying ¬¬(X ∨ ¬X), given that X is some type in Pi.

```scheme
; (X ∨ ¬X)

(claim pem
  (Pi ((X U))
    (Either X (-> X Absurd))))

; ¬¬(X ∨ ¬X)

(claim pem-not-false
  (Pi ((X U))
    (-> (-> (Either X (-> X Absurd))
            Absurd)
        Absurd)))
(define pem-not-false
  (lambda (X)
    (lambda (pem-false)
      (pem-false
        (right
          (lambda (x)
            (pem-false
              (left x))))))))
```

Thus, we are left in the situation of proving `pem-not-false` but not being able to construct a program for `pem`

`¯\_(ツ)_/¯`

[top](#)
