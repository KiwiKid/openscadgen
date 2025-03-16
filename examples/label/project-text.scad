include <BOSL2/std.scad>

$fn = 20;

name = "Rodney Cumming";

text_angle = 60;

base_size = [130, 30, 20];

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