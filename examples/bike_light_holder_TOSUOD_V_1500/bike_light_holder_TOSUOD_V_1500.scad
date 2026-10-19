include <BOSL2/std.scad>;

$fn = 100;


bodyWidth = 50; 
bodyHeight = 60;

bodyLength = 80;
bodyRounding = 5;

torchRounding = 3;
torchRotation = 5;

torchCutoutUp = 15;

torchSize = [34,100,30];
torchBottomPlacementCube = [4,5,5];

bikeFrameConnectionSize = [35,120,50];
bikeFrameConnectionRounding = 10;
bikeFrameConnectionDown = 0;
bikeFrameConnectionRadius = 20;



bikeTopFrameConnectionSize = [40,120,50];
bikeTopFrameConnectionRounding = 17;
bikeTopFrameConnectionRadius = 20;



module tieSlot(tieSize=[10,10], h=20, anchor=BOT){
    up(torchCutoutUp)

    xrot(90)
    rect_tube(wall= 6, h = h, size = tieSize, rounding=15, anchor=anchor);
}

module torchCutout(){
up(torchCutoutUp)
difference(){
union(){
    cuboid(torchSize, rounding=torchRounding, anchor=BOT);

        back(30)
tieSlot(tieSize=[torchSize[0]+18, torchSize[2]+18], anchor=BOT);

    fwd(10)
tieSlot(tieSize=[torchSize[0]+18, torchSize[2]+18], anchor=BOT);
    }
    
    
    cuboid(torchBottomPlacementCube, rounding=0.5, anchor=BOT, edges=TOP);
    
    
    }
}

module bikeTopFrameConnection(){
down(bikeFrameConnectionDown){
    
        fwd(bodyLength/2-bikeTopFrameConnectionSize[2])
        zrot(90)
       cuboid(bikeTopFrameConnectionSize, rounding=bikeFrameConnectionRounding, anchor=TOP, edges="Y");
       
       down(40)
    
    fwd(30)
    tieSlot(tieSize=[bikeTopFrameConnectionSize[0]+18, bikeTopFrameConnectionSize[2]+18], anchor=TOP);
    }
}

module bikeFrameConnection(){
down(bikeFrameConnectionDown)
    
    
       cuboid(bikeFrameConnectionSize, rounding=bikeTopFrameConnectionRounding, anchor=TOP, edges="Y");
}


module bike_light_holder_TOSUOD_V_1500(){
difference(){
	cuboid([bodyWidth,bodyLength,bodyHeight], anchor=CENTER, rounding=bodyRounding);
    
   xrot(torchRotation)
    torchCutout();
    
    
    bikeFrameConnection();
    
    
    bikeTopFrameConnection();
    }
}

//front_half()
bike_light_holder_TOSUOD_V_1500();
