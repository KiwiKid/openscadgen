include <BOSL2/std.scad>
include <BOSL2/attachments.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

cyliderDiamter = 70;

bottom_cube=900;
cubeRotate=25;
cubeTranslate=[-120,-40,0];

side_thru_cutout_diameter = 15;

cutout_down = 50;

holder_cutour_diameter = 25;

module hex_with_cylinder() {
    // Create the hex prism and define an attachment point on top
    attachable(){
    up(2)
    tube(or=11, ir=1, h=23, $fn=6, rounding=1.3, teardrop=true, anchor=TOP);

    difference(){
    
    cube_height = 60;
    cube_width = 70;
        up(cube_height/2)
        cuboid(size=[cube_width,cube_width, cube_height], rounding=3);
        //rect_tube(d=70, h=40, anchor=BOTTOM);
        
        // Handle cutout
        rotate([0,90,0])
        left(cutout_down-20)
        cylinder(d=holder_cutour_diameter, h=300, anchor=BOTTOM);
        
        // Handle cutout extended
        rotate([0,90,0])
        left(cutout_down+10)
        cuboid([60,holder_cutour_diameter*0.7,300], anchor=BOTTOM);
        
        // Side thru cutout
      /*  rotate([90,90,0])
        left(cutout_down-25)
        down(cutout_down)
        cylinder(d=side_thru_cutout_diameter, h=300, anchor=BOTTOM);
        
        // Side thru cutout extender
        rotate([90,90,0])
        left(cutout_down)
        cuboid([50,side_thru_cutout_diameter,100], anchor=BOTTOM);
        */
        
        // Bottom cutout
        up(5)
        cyl(d=45, h=60, anchor=BOTTOM, rounding1=10);
         
       /*  translate(cubeTranslate)
         rotate([0,-cubeRotate,0])
         cuboid(bottom_cube);*/
        }
    }
}

// Render the shape
hex_with_cylinder();
