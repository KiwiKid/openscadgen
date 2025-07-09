regular_prism_n = 5;
regular_prism_h = 25;
regular_prism_r = 10;
regular_prism_center = true;
regular_prism_realign = false;
regular_prism_shift = [0,0];
regular_prism_chamfer = 0;
regular_prism_chamfer1 = 0;
regular_prism_chamfer2 = 0;
regular_prism_rounding = 0;
regular_prism_rounding1 = 0;
regular_prism_rounding2 = 0;
regular_prism_anchor = "CENTER";
regular_prism_spin = 0;
regular_prism_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module regular_prism_demo(){
	regular_prism(n=regular_prism_n, h=regular_prism_h, r=regular_prism_r, center=regular_prism_center, realign=regular_prism_realign, shift=regular_prism_shift, chamfer=regular_prism_chamfer, chamfer1=regular_prism_chamfer1, chamfer2=regular_prism_chamfer2, rounding=regular_prism_rounding, rounding1=regular_prism_rounding1, rounding2=regular_prism_rounding2, anchor=regular_prism_anchor, spin=regular_prism_spin, orient=regular_prism_orient);
}

regular_prism_demo();
