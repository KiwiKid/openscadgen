include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

thingFrameRounding = 2;
insetThing = "square"; // square round


/// Frame size
frameWallWidth = 5;
frameDepth = 5;

insetPrimarySize = 49;
insetSecondarySize = 49;

// FACEPLATE STYLE
frameWidth = insetPrimarySize + frameWallWidth;
frameHeight = insetSecondarySize + frameWallWidth;
squareFrameSize = [frameDepth,frameWidth,frameHeight];


insetDepth = 6;

insetOffset = frameDepth/2-insetDepth/2+0.01;

// square inset settings
squareInsetSize = [insetDepth, insetPrimarySize, insetSecondarySize];
insetRounding = 8;
smallHookMove = [10, -5,0];
largeHookMove = [15, -5, 0];

// round settings

roundHolderSize = 22;
roundInsertRadius = 20;
roundInsertSize = 20;
roundInsertDepth = 10;

/// PICTURE-RAIL HOOK
attachmentType = "floorPanel"; // "railHook" or "poll" or "floorPanel"

// The rail runs along Y.  Adjust these three measurements to suit the rail.
railDepth = 18;
railHeight = 7;
railClearance = 2;
railHookWall = 2;
smallHookRoundMove = [0, 0, 0];

// POLL ATTACHMENT
// A cord channel running along Y.  The underside slot lets the cord snap in.
pollDiameter = 10;
pollWall = 2;
pollSlotWidth = 3;
    
                usbHeight = 25;
            usbWidth = 20;
             usbDepth = 10;
             
             floorPanelLength = 30;
             floorPanelSize = [floorPanelLength,25*2,2];
             frameAttachAdjust = [floorPanelLength/6+2,0,-2];
             frameAttachScrewAdjust = [floorPanelLength/2,0,0];
             
pollHeight = 170;

// An inverted J: the holder is the front leg, with a bridge over the rail and
// a short back leg that prevents it lifting away from the rail.
module picture_rail_hook() {
    bridgeDepth = railDepth + railClearance + 2 * railHookWall;
    backLegX = frameDepth / 2 + railDepth + railClearance;

    // Overlap the holder at its top/right edge so both solids are joined.
    translate([frameDepth / 2 - railHookWall, -frameWidth / 2, frameHeight / 2 - railHookWall])
        cuboid([bridgeDepth, frameWidth, railHookWall], anchor=LEFT+FRONT+BOTTOM, rounding=1);

    // Retaining leg behind the rail.
    translate([backLegX, -frameWidth / 2, frameHeight / 2 - railHeight - railHookWall])
        cuboid([railHookWall, frameWidth, railHeight + railHookWall], anchor=LEFT+FRONT+BOTTOM, rounding=1);
}

module floor_panel_panel(){
difference(){
    move(frameAttachAdjust)
cuboid(floorPanelSize, anchor=BOT, rounding=0.5, except_edges=TOP);

    move(frameAttachScrewAdjust)
    #zcyl(r=2,h=10);
}

}

module floor_panel_frame(frameAngle=15){
    #holder()
    yrot(frameAngle)
    attach(BOT)
    floor_panel_panel();

}

module poll_attachment(
    pollHeight = pollHeight, usbDepth=usbDepth,usbWidth=usbWidth, usbHeight=usbHeight, pollPart="top") {
    pollRadius = pollDiameter / 2;
    pollCenter = [
        frameDepth / 2 + pollRadius /2 - pollWall,
        0,
        -frameHeight - pollWall + 4
    ];
    
    
    if(pollPart=="top"){
    
    
    }else{
    

    difference() {
        translate(pollCenter)
            rotate([90, 0, 0])
                ycyl(d=pollDiameter, h=frameWidth, chamfer1=-30, chamfer2=-10);

        translate(pollCenter)
            rotate([90, 0, 0])
                ycyl(d=pollDiameter - 2 * pollWall, h=frameWidth + 3);

        // Open the underside just enough for the cord to be pressed into the tube.
        rotate([0,0,-90])
        translate([-pollSlotWidth/2 , + 3, pollCenter.z - pollHeight])
            cube([pollSlotWidth, pollRadius + 1, frameWidth + 2]);
            
            down(frameHeight/2 + pollHeight /2 - usbHeight + pollWall *2 + 2 )
            left(usbDepth/2-2)
            fwd(usbWidth/2)
            cube([usbDepth,usbWidth,usbHeight]);
    }
    }
             

             
          
}

module attachment(type=attachmentType) {
    if (type == "railHook") {
        picture_rail_hook();
    } else if (type == "poll-top") {

        poll_attachment(pollPart="top");
    } else if (type == "poll-bottom") {

        poll_attachment(pollPart="bottom");
    } else if(type == "floorPanel") {
        floor_panel_frame();
    }else {
        assert(false, str("Unknown attachmentType: ", type));
    }
}

module holder(insetThing="square"){
    if(insetThing == "square"){
        cuboid(squareFrameSize, rounding=thingFrameRounding) children();
    } else if(insetThing == "round"){
        rotate([0,90,0])
        cyl(r=roundHolderSize, h=frameDepth, rounding=1)
         rotate([0,-90,0])
         move(smallHookRoundMove)
         children();
    } else {
        echo("!!!!!!!!! insetThing is not valid");
    }
}

module any_holder(){
intersection(){
    difference(){
        // GROUP 1: The solid positive shapes to be joined together
        // Enclosing these ensures the holder + attachment pass through cleanly
        union() {
            holder(insetThing=insetThing)
                attachment();
        }
        
        // GROUP 2: The cutters (everything below this point gets subtracted)
        if(insetThing == "square"){
            left(insetOffset)
                cuboid(squareInsetSize, rounding=insetRounding, edges="X");
                
                // cord hole
                down(frameWidth/2)
                right(2)
                cuboid([5,15,10]);
                
        } else if(insetThing == "round"){
        
           rotate([0,90,0])
           //down(roundInsertDepth)
                cyl(r=roundInsertRadius, h=insetDepth+3, rounding1=-2);
                
                  // cord hole
                //down(roundInsertSize)
                //right(-roundInsertDepth*1.3)
                //#cuboid([5,15,10]);
                
        } else {
            echo("!!!!!!!!! insetThing is not valid");
        }
    }
    intersetBoxOffset = 0;
        intersetBoxLeftShift = -.7;
    union(){
    // Bounding cuboid
        //    down(frameHeight/2 + pollHeight/2 - usbHeight*2+1)
         //   left(usbDepth/2-2)
          //  fwd(usbWidth/2)
         left(intersetBoxLeftShift)
          down(frameHeight+intersetBoxOffset)
            cuboid([15,15,frameHeight], rounding=2, anchor=LEFT);
            
            down(intersetBoxOffset)
         left(intersetBoxLeftShift)
            cuboid([25,80,frameHeight], rounding=2, anchor=LEFT);
            
            }
            }
    
}

any_holder();
