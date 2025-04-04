include <BOSL2/std.scad>

$fn =20;

    r1 = 25; r2 = 12; R = 65;
  length = 60;

  height = 50;
  rounded = "true";
    
module carovan_hook_shape(){ 


        
        
         connector_move = [5,28.4,0];
     
     connector_size = [65,20];
     difference_move = [0,37,0];
     difference_size = [-2,-5];
     
     difference(){
     union(){
        rotate([0,0,-40])
        difference(){
            egg(length,r1,r2,R,$fn=180);
            move([2,11,0])
           egg(60,r1,r2,R,$fn=180);
         }
         move(connector_move)
        rect(connector_size, 2);
    }

        move(difference_move)
        rect(connector_size-difference_size);
    }

    holder_size = 35;
    holder_width = 30;
    holder_move = [50,48,0];
    holder_cutout_move = [-23,0,0];

    holder_cutout_size = [70,23];

    holder_cutout_2_move = [10,10,0];

    move(holder_move)
    difference(){

        rect([holder_size,holder_width], 2);


        move(holder_cutout_move)
        rect(holder_cutout_size);
        
        left(15)
        back(10)
        move(holder_cutout_move)
        rect(holder_cutout_size);
        
    }

    }
  
if(rounded == "true"){

    corner_radius = 2;
    minkowski() {
        scale(.7)
        linear_extrude(height=height)
        carovan_hook_shape();
        
        // Small sphere or cylinder to round corners
        sphere(r=corner_radius, $fn=$fn);
    }
    }else{
            carovan_hook_shape();
}
    
    
    carovan_hook_shape();
