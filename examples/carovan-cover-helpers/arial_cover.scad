include <BOSL2/std.scad>
include <BOSL2/joiners.scad>
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

arial_height = 250;
arial_smaller_bottom_radis = 130;

// Parameters
arc_radius = 254;      // Radius of the arc
arc_angle = 90;       // Arc angle in degrees
cutout_width = 130;    // Width of the rectangle cutout
cutout_height = 200;   // Height of the rectangle cutout
cutout_translate = [cutout_width/2, arial_height, 0];

// Module for the filled arc using polygon
module arc_shape(radius, angle) {
    points = concat(
        [[0, 0]],
        arc(r=radius, angle=angle)
    ); 
    polygon(points);
}

// Module for the cutout
module cutout(width, height) {
        square([width, height], anchor=CENTER);
        rotate([0,0,20])
        translate([-45,-17,0])
        square([width+10, height+3], anchor=CENTER);
}

// Combine arc and cutout
module final_shape() {
    
    translate([arial_smaller_bottom_radis, 0,0])
    difference() {
        arc_shape(arc_radius, arc_angle);
         translate(cutout_translate)
        cutout(cutout_width, cutout_height);
    }
}


thin_extrude = true;  // Set to false for full rotate_extrude

if (thin_extrude) {
    linear_extrude(height=0.2) final_shape();  // Thin extrusion (plate-like)
} else {
    difference(){
        rotate_extrude() final_shape();;  // Full revolve (3D shape)
        #dovetail("male", width=30, height=8, slide=30);
    }
}
