cuboid_size = [100,100,100];
cuboid_chamfer = 0;
cuboid_rounding = 0;
cuboid_edges = "EDGES_ALL";
cuboid_except = [];
cuboid_trimcorners = true;
cuboid_teardrop = false;
cuboid_anchor = "CENTER";
cuboid_spin = 0;
cuboid_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module cuboid_demo(){
	cuboid(size=cuboid_size, chamfer=cuboid_chamfer, rounding=cuboid_rounding, edges=cuboid_edges, except=cuboid_except, trimcorners=cuboid_trimcorners, teardrop=cuboid_teardrop, anchor=cuboid_anchor, spin=cuboid_spin, orient=cuboid_orient);
}

cuboid_demo();
