include <BOSL2/std.scad>



module carovan_hook_shape(){ 

r1 = 25; r2 = 12; R = 65;
    length = 70;
    
    
     connector_move = [15,28.4,0];
 
 connector_size = [75,15];
 difference_move = [10,35,0];
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

holder_size = 40;
holder_width = 30;
holder_move = [68.5,48,0];
holder_cutout_move = [-10,0,0];

holder_cutout_size = [42,20];

holder_cutout_2_move = [-50,10,0];

move(holder_move)
difference(){

    rect([holder_size,holder_width], 2);


    move(holder_cutout_move)
    rect(holder_cutout_size);
    
}

}
height = 1;
scale(.7)
linear_extrude(height=height)
carovan_hook_shape();