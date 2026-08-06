
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


wallDepth = 1;
radius = 30;
height = 45;

bowlHoles = "prismoidHoles"; // none
holeOffset = 9;
holeSizeRatio = 8;

module bowlShape(radius=radius){
cyl(r=radius, h=height, anchor=BOTTOM, chamfer1=radius/5);

}

module holeSet(){
up(height/2+2)
            rot_copies(n=8, cp=[0, 0],delta=[radius-holeOffset,0,0]) 
            
            yrot(90)
            zrot(45)
            prismoid(size1=[height/holeSizeRatio+radius/holeSizeRatio,height/holeSizeRatio+radius/holeSizeRatio], size2=[height/holeSizeRatio+radius/holeSizeRatio,height/holeSizeRatio+radius/holeSizeRatio], h=14);
            }

module old_light_switch_cover_smart_light(bowlHoles="prismoidHoles"){
	
    difference(){
    bowlShape(radius=radius);
    up(wallDepth)
        bowlShape(radius=radius-wallDepth);
        
        
        if(bowlHoles == "prismoidHoles"){
            holeSet();
            
            up(10)
            zrot(360/8/2)
            holeSet();
            
            
            down(10)
            zrot(360/8/2)
            holeSet();
        }
        }
    

    
//        yrot(180)
        //chamfer_cylinder_mask(r=radius, chamfer=1);
}


old_light_switch_cover_smart_light(bowlHoles=bowlHoles);
