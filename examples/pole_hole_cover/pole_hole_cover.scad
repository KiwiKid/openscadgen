include <BOSL2/std.scad>;

$fn = 100;

mode="topModifer"; // "topModifer" plug

plugSize = [160, 160, 20];


movePlugUpAmount = 12;
movePlugUp = [0,0,movePlugUpAmount];
topPatternSize = 7;
topPatternHeight = 1;

topPatternBaseHeight = 1;
smileAmountAdjust = 2.7;

globalBounds=  [300,300,150];
globalBoundsMove = [0,0,-11];
stlScale = [1.4,1.4,1];


module topPatternRaw(mode="plug"){
 if(mode != "onlyModifer"){
        color("gold")
        cylinder(r = topPatternSize, h = topPatternBaseHeight, anchor=BOT)
       ;
    }

    // Left Eye
    color("black")
    translate([-topPatternSize/2.7, topPatternSize/3.1, 0.001])
    cyl(r = topPatternSize/8.3, h = topPatternHeight, anchor=BOT);

    // Right Eye
    color("black")
    translate([topPatternSize/2.7, topPatternSize/3.1, 0.001])
    cyl(r = topPatternSize/8.3, h = topPatternHeight, anchor=BOT);

    // Smile
    color("black")
    translate([0, -2, 0.001])
    difference() {
        // Outer arc radius
        cyl(r = topPatternSize/1.78, h = topPatternHeight, anchor=BOT);
        // Inner cutout to create line thickness
        //down(0.001)
       // cyl(r = topPatternSize/2.27, h = topPatternHeight, anchor=BOT);
        // Mask off top half to leave only the bottom curve
        down(0.001)
        translate([0, topPatternSize*1.4, 0])
        cuboid([30, topPatternSize*smileAmountAdjust, topPatternSize], anchor=BOT);
    }
}


module topPattern(mode="onlyModifer"){
    topPatternRaw(mode=mode);
}

module anchored_stl(stl_path, size, anchor=CENTER, spin=0, orient=UP, stlMoveAdjust=stlMoveAdjust, globalBounds=globalBounds, globalBoundsMove=globalBoundsMove, stlScale=stlScale) {
move(stlMoveAdjust)
scale(stlScale)
intersection(){
    attachable(anchor=anchor, spin=spin, orient=orient, size=size) {
        // Center the imported mesh at local [0,0,0]
        import(stl_path, center=false);
        children();
        
        
    }
    move(globalBoundsMove)
    cuboid(globalBounds, anchor=TOP);
    }
}


module pole_hole_cover(mode="plug", size=plugSize){
echo(str("====== ", mode, "======"));

    if(mode == "plug"){
//	resize(plugSize) {
        anchored_stl("./screw_hole_cover.stl", size=plugSize, stlScale=[], anchor=TOP, stlMoveAdjust=movePlugUp);

 //   }
    } else if(mode == "topModifer"){
   // up(plugSize[2])
     //   resize(plugSize){
        topPattern(mode="onlyModifer");
  //  }
    }else if(mode == "all"){
     #anchored_stl("./screw_hole_cover.stl", size=plugSize, anchor=TOP, stlMoveAdjust=movePlugUp);
     topPattern(mode="pattern");
    } else {
    assert(false, "invalid mode");
    }
}

pole_hole_cover(mode=mode, size=plugSize);


