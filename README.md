# goors
The goal of this project is to implement an almost state-of-the-art 2D orthogonal range searching structure.

The current implementation is reminiscent of range trees with fractional cascading.
We can currently query the structure in O(log n + k) time, where k is the number of outputs and it uses O(n log n) words of storage.

The implementation is also reminiscent of [this paper](http://cs.au.dk/~larsen/papers/orth_revisit.pdf). In fact here we have simplified that structure slightly.

The goal of the paper above is to provide a linear space solution with efficient query times.
Specifically the query time when going to linear space becomes roughly O(log^epsilon n) for any epsilon > 0.
The paper presents a reduction from 2D orthogonal range searching to what they call the 'Ball Inheritance Problem'.

## Idea
The main idea for 2D orthogonal range search is as follows:
Build a balanced tree on the sorted x-coordinates of the points.
Now a subtree rooted at a node N has a subset of the points (assume we only stores values in the leaves).
Take all the points in that subtree, and build a bit-vector at node N.
To build the bit-vector, take the points in sorted y-order, and see for each point if it falls in the left subtree (corresponding to a 0) or the right (corresponding to a 1).

## Answering a query
To answer a query range [x0, x1] x [y0, y1], we find the two leaves corresponding to x0 and x1 and find their lowest common ancestor (lca).
Descend from the lca down to the leaves and report subtrees that hang 'in-ward'.
However, we have not taken into account [y0, y1].
The ranks y0 and y1 (index in sorted list of y-coordinates if inserted), correspond to an interval of bits in the root node, which are the y-coordinates that fall in the range [y0, y1].
Note that these can be maintained as we descend to children if we can efficiently count the number of 1s up to every position in the bit-vectors.
That operation is what the gorasp project implements -- it is called RankOfIndex.

So to report an 'in-ward' subtree (i.e. report all points in it) we only report those with the right y-coordinates.
That means we need to be able to find out at a node for an index in the bit vector, which leaf in this subtree would it eventually reach?
Answering that question is what the Ball Inheritance problem is all about.
And that is also where the entire trade-off in the solution is to be found.
One way, is to just store a pointer for every index. This requires O(n log n) pointers in total -- current solution, and gives O(1) time per point reported.
Alternatively one could not store anything, and just follow the bit vectors down to the leaf, required O(log n) per point, but maintaining overall O(n) space.

# Remarks
If I had more time, I would have seperated out the ball inheritance to stand on its own, currently it sits in the implementation of the tree, which is inconvenient if we want to use other strategies.
It is also something that can change all on its own, so we should (following object oriented practices) seperate it out, give an interface and provide implementations.
Perhaps one could also argue, that it would then be appropriate to also have factory for instantiating a range searching structure with the desired trade-offs.

## Optimization
The implementation uses an implicit tree representation.
From algorithm engineering litterature we found out that laying out a tree in BFS order usually gives decent cache performance, and always better than a pointer-based structure.
So that is why I went with that representation.
It makes the code harder to read and therefore understand than the alternative pointer-based structure.

# Speed
On my machine (3.2ghz) the structure can answer about 1600 queries/second for around 75000 points.
If I had more time I would like to benchmark and see how it compares to naive things such as just scanning all points and testing if it should be reported.
If I had even more time, I would like to implement a kd-tree which is also known to perform well in practice, however it has worse guarantess on the running time.
A kd-tree gives O(sqrt(n) + k) query time (and so does quad-trees and the variations of that theme).
