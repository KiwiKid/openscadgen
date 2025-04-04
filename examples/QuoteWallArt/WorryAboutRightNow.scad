include <BOSL2/std.scad>

$fn = 20;

name = "The biggest thing we have to worry about right now";
name2 = "is what's going on right now.";
name3 = "The rest will take care of itself.";

text_size = 6;
text_angle = 60;
text_internal_radus = 500;
text_external_radus = 590;

text_down = 400;


base_height = 15;

 base_width = !is_undef(base_width) ? base_width : 178 ;
base_size = [base_width, 10, base_height];


base_z_offset = 480;
base_y_offset = 280;

base_y_offset_2 = 295;
base_size_2 = [108, 25, base_height];

diff_z_offset = 46;

font = "Oswald";

difference(){
union(){
rotate([-text_angle,0,0])
cylindrical_extrude(or=text_external_radus, ir=text_internal_radus)
 text(text=name, size=text_size, halign="center", valign="center", font=font);
 
 rotate([-text_angle,0,0])
cylindrical_extrude(or=text_external_radus, ir=text_internal_radus)
  fwd(10)
  text(text=name2, size=text_size, halign="center", valign="center", font=font);
  
  
  rotate([-text_angle,0,0])
cylindrical_extrude(or=text_external_radus, ir=text_internal_radus)
  fwd(20)
 text(text=name3, size=text_size, halign="center", valign="center", font=font);

 
 
    fwd(base_y_offset)
    up(base_z_offset)
     cuboid(base_size, rounding=2);
     
         fwd(base_y_offset_2)
    up(base_z_offset)
    cuboid(base_size_2, rounding=2);
 }
 
 fwd(base_y_offset)
 up(base_z_offset-diff_z_offset)
    cuboid([305,305,100]);
    
    
    
    
    
  //  rotate([text_angle,0,0])
   //    #cuboid([300,300,10]);

 }
 
  fwd(base_y_offset-10)
  up(base_z_offset)
 difference(){
  rotate([180,0,0])
 teardrop2d(r=12, ang=30, cap_h=40);
 down(10)
 #cylinder(100, 7,7);
 }