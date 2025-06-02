include <BOSL2/std.scad>

$fn = 20;

name = "Your Name Here";

text_angle = 60;
base_depth = 25;
base_height = 15;
 base_width = !is_undef(base_width) ? base_width : 130 ;
base_size = [base_width, base_depth, base_height];

difference(){
union(){



rotate([-text_angle,0,0])
cylindrical_extrude(or=140, ir=110)
 text(text=name, size=12, halign="center", valign="center");
 
 
 
    fwd(60)
    up(93)
     cuboid(base_size, rounding=2);
 }
 
 up(84)
    cuboid([300,300,30]);
    
    
  //  rotate([text_angle,0,0])
   //    #cuboid([300,300,10]);

 }