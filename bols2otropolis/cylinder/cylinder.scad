cylinder_h = 50;
cylinder_r = 25;
cylinder_center = true;
cylinder_chamfer = 0;
cylinder_rounding = 0;
cylinder_anchor = "CENTER";
cylinder_spin = 0;
cylinder_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module cylinder_demo(){
	cylinder(h=cylinder_h, r=cylinder_r, center=cylinder_center, chamfer=cylinder_chamfer, rounding=cylinder_rounding, anchor=cylinder_anchor, spin=cylinder_spin, orient=cylinder_orient);
}

cylinder_demo(); 