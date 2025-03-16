include <BOSL2/std.scad>



module carovan_hook_shape(){ 

r1 = 25; r2 = 12; R = 65;
    length = 70;
    
    
     connector_move = [28,21.4,0];
 
 connector_size = [100,20];
 difference_move = [5,25,0];
 difference_size = [10,5];
 
 difference(){
 union(){
 
rotate([0,0,-40])
difference(){

    egg(length,r1,r2,R,$fn=180);

    move([2,8,0])
   egg(60,r1,r2,R,$fn=180);
    
 }
 
 move(connector_move)
rect(connector_size, 2);
}



 
 


move(difference_move)
rect(connector_size-difference_size, 2);


}

holder_size = 100;
holder_width = 40;
holder_move = [120,40,0];
holder_cutout_move = [-10,0,0];

move(holder_move)
difference(){

    rect([holder_size,holder_width], 2);


    move(holder_cutout_move)
    #rect([holder_size-10,holder_width-10], 2);

}

}


linear_extrude(height=10)
carovan_hook_shape();