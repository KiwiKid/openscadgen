include <BOSL2/std.scad>
include <BOSL2/isosurface.scad>
$fn = 200;


circleDiameter = 10;

toggleHeight = 10;

cutout_size = [30,5,10];
cutout_tranlate = [-30,0,0];

toggle_connection_translate = [0,0,2];
toggle_connection_size = [100,40,10];

toggle_connector_cuboid_translate = [10, 0, 0];
toggle_connector_cuboid_size = [50, circleDiameter*2-5, 5];

toggle_height = 10;
toggle_radius = 5;
toggle_translate = [6,0,0];

toggle_gap_cuboid_translate = [20, 0, -15];
toggle_gap_cuboid_size = [40, toggle_radius*2+1, 30];



module toggle(){
    difference(){
        translate(toggle_connection_translate)
        cuboid(toggle_connection_size);
        double_circle(scaleY=1.05, scaleX=1.05, scaleZ=1.05, type="female");
        
        translate(toggle_gap_cuboid_translate)
        cuboid(toggle_gap_cuboid_size, anchor=BOTTOM, rounding=5);
        
        // Bottom flattener
        translate([0,0,-8])
        cuboid(toggle_connection_size);
     }
     
    difference(){
    union(){
        double_circle(type="male");
        //translate(toggle_connector_cuboid_translate)
        //#cuboid(toggle_connector_cuboid_size);
        
        translate(toggle_translate)
        cylinder(toggle_height, toggle_radius,toggle_radius);
        }
            translate(cutout_tranlate)
            cuboid(cutout_size, anchor=CENTER);
            
            
           // Bottom flattener

        translate([0,0,-8])
        cuboid(toggle_connection_size);
    }
   

}

disk_height = 9;
disk_radius = 6;

mbs = [];
module double_circle(scaleX = 1, scaleY = 1, scaleZ = 1, type = "male", includeFourth = false){
if(type == "female"){
      scale([scaleX, scaleY, scaleZ])
    metaballs([
       move([-20, 0, 0]), mb_cyl(disk_radius, disk_height),
    move([0, 0, 0]), mb_cyl(disk_radius, disk_height),
    move([20, 0, 0]), mb_cyl(disk_radius, disk_height),
    move([40, 0, 0]), mb_cyl(disk_radius, disk_height) 
        ], [[-30,-10,-3], [30,10,3]], 0.5);
    }else{
       
    
  
      scale([scaleX, scaleY, scaleZ])
    metaballs([
        move([-20,0,0]), mb_cyl(disk_radius,disk_height),
        move([0,0,0]), mb_cyl(disk_radius,disk_height),
       // move([20,0,0]), mb_disk(disk_radius,disk_height)
        ], [[-30,-10,-3], [30,10,3]], 0.5);
  }
  }
  
  
  scale([0.3,0.3,0.3])
  toggle();