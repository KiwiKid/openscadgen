

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 20;
    
handleCoverHeight = 120;
radius1 = 40;
radius2 = 50;
arcRadius = 300;
intersect = true;

gapMove = [arcRadius-160,0,0];
module door_handle_cover(){
    difference(){
cyl(r1=radius1, r2=radius2, h=handleCoverHeight, rounding=30);
    
    xrot(90)
    yrot(90)
    move(gapMove)
   path_extrude2d(arc(d=arcRadius,angle=[160,200], n = 100),caps=true)
    trapezoid(w1=20, w2=10, h=70, anchor=BACK);
    }
}



if(intersect){
   intersection(){
   
door_handle_cover();
left(20)
cuboid([1.3,250,250]);
}
   }else {

door_handle_cover();
}
