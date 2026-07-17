
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


wallDepth = 1;
radius = 30;
height = 30;

bowlHoles = "prismoidHoles"; // none
holeOffset = 2;

module bowlShape(radius=radius){
cyl(r=radius, h=height, anchor=BOTTOM, chamfer1=radius/5);

}

module old_light_switch_cover_smart_light(bowlHoles="prismoidHoles"){
	
    difference(){
    bowlShape(radius=radius);
    up(wallDepth)
        bowlShape(radius=radius-wallDepth);
        
        
        if(bowlHoles == "prismoidHoles"){
        up(height/2+2)
            rot_copies(n=8, cp=[0, 0],delta=[radius-holeOffset,0,0]) 
            
            yrot(90)
            zrot(45)
            prismoid(size1=[13,13], size2=[0,0], h=14);
        }
        }
    

    
//        yrot(180)
        //chamfer_cylinder_mask(r=radius, chamfer=1);
}


old_light_switch_cover_smart_light(bowlHoles=bowlHoles);
