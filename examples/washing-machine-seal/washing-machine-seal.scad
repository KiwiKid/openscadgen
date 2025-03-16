include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


module washing_machine_seal(){

    // Adjustable parameters
    seal_thickness = 10;
    outer_width = 6;
    inner_width = 12;
    seal_height = 25;
    seal_rim_z_offset = -20;
    seal_rim_x_offset = 5;

    seal_rim_cutout_width = 25;

    // Profile shape
    module seal_profile() {
        difference() {
            // Outer block of seal
            rect([outer_width, seal_height], anchor=TOP, rounding=0.3);
            // Cutout for the U-channel
            move([outer_width/2-inner_width/2+seal_rim_x_offset, seal_thickness+seal_rim_z_offset])
                rect([inner_width+seal_rim_cutout_width, seal_height-seal_thickness+6], anchor=RIGHT, rounding=0.3);
                
               move([0,-30,0])
               rotate([0,0,40])
            rect([10,20]);
        }
    }

    // Define a curved path
    path = arc(r=170, angle=120, n=$fn);  // Adjust curvature

    difference(){
    path_extrude(path=path) {
        seal_profile();
    }
    
    up(seal_height-4)
    rotate([0,0,-30])
       cuboid([500,270,3]);
    }

}




  renderType = "";
  
  if(renderType == "horz-slice"){
      intersection(){
      washing_machine_seal(); 
      fwd(500)
      left(500)
      down(-25)
      cube([1000,1000,0.3]);
      }
  }else if(renderType =="vert-slice"){
       intersection(){
          #washing_machine_seal();
          rotate([90,0,90])
          fwd(500)
          rotate([0,-45,0])
          cube([1000,1000,0.3]);
      }
  }else{ 
  
  washing_machine_seal();
  }