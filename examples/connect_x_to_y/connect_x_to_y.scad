include <BOSL2/std.scad>;

$fn = 100;

xShape = "tube";
yShape = "tube";

tubeChamfer1 = 3;

xHeight = 60;
yHeight = 80;
tubeRadius = 26.2/2;
tubeConnectorRadius = 35/2;
tubeConnectorHeight = 10; // 120;
tubeConnectorOverlap = 36;
tubeWall = 3;



height = 10;

enableY = true;
enableX = true;

globalBounds = [90,90,90];
globalBoundsMove = [0,0,20];
globalBoundsRotation = [0,30,50];

module x(shape="tube"){
    if(shape == "tube"){
    
        tube(or=tubeRadius, wall=tubeWall, h=xHeight, anchor=BOT);
    } else {
        assert(false, "in x, shape not supported");
    }
}
module y(shape="tube"){
 if(shape == "tube"){
   tube(or=tubeRadius, wall=tubeWall, h=yHeight, anchor=TOP, ichamfer2=tubeChamfer1);
    } else {
        assert(false, "in x, shape not supported");
    }
}
module connector(){
    tube(or=tubeConnectorRadius, wall=tubeConnectorRadius-tubeRadius, h=tubeConnectorHeight, anchor=TOP);
    //cyl(r=tubeConnectorRadius, h=tubeConnectorHeight, anchor=TOP);
}
module connect_x_to_y(){
if(enableX){
    down(tubeConnectorOverlap)
	x(shape=xShape);
    }
    connector();
    
    down(tubeConnectorHeight-tubeConnectorOverlap)
    if(enableY){
    	y(shape=yShape);
        }
}


intersection(){
move(globalBoundsMove)
rotate(globalBoundsRotation)
#cuboid(globalBounds);
connect_x_to_y();
}
