
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

matchHandleHeight = 40;
matchHandleDiameterBot = 5;
matchHandleDiameterTop = 5;


matchSlotWidth = 2.5;
matchSlotDepth = 2.5;

module match_handle(matchHandleHeight=matchHandleHeight, matchHandleDiameter=matchHandleDiameter, matchSlotWidth=matchSlotWidth, matchSlotDepth=matchSlotDepth, matchHandleDiameterTop=matchHandleDiameterTop, matchHandleDiameterBot=matchHandleDiameterBot){
    
    down(matchHandleHeight/2)
cyl(h=10, d=10);

	difference(){
		cyl(h=matchHandleHeight, d1=matchHandleDiameterBot, d2=matchHandleDiameterTop);
        up(0.01)
		cuboid([matchSlotWidth,matchSlotDepth,matchHandleHeight], chamfer=-.5);
	}
	
}


match_handle();
