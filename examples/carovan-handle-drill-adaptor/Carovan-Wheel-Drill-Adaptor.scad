include <BOSL2/std.scad>
include <BOSL2/attachments.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

bottom_cube=60;
side_thru_cutout_diameter = 15;

holder_cutour_diameter = 25;

module hex_with_cylinder() {
    // Create the hex prism and define an attachment point on top
    attachable(){
    up(2)
    tube(or=11, ir=1, h=15, $fn=6, rounding=1.3, teardrop=true, anchor=TOP);

    difference(){
    
        cylinder(d=50, h=50, anchor=BOTTOM);
        
        // Handle cutout
        rotate([0,90,0])
        left(16)
        cylinder(d=holder_cutour_diameter, h=30, anchor=BOTTOM);
        
        // Handle cutout extended
        rotate([0,90,0])
        left(27)
        cuboid([30,holder_cutour_diameter,30], anchor=BOTTOM);
        
        // Side thru cutout
        rotate([90,90,0])
        left(15)
        down(30)
        cylinder(d=side_thru_cutout_diameter, h=60, anchor=BOTTOM);
        
        // Side thru cutout extender
        rotate([90,90,0])
        left(38)
        down(30)
        cube([50,side_thru_cutout_diameter,100], anchor=BOTTOM);
        
        
        // Bottom cutout
        up(5)
        #cyl(d=38, h=30, anchor=BOTTOM, rounding1=10);
         
         fwd(bottom_cube/2)
         up(25)
         left(bottom_cube/2)
         rotate([0,-10,0])
         cube(bottom_cube);
        }
    }
}

// Render the shape
hex_with_cylinder();
