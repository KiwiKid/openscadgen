
// 1. Library Includes
include <BOSL2/std.scad>

// 2. Render Quality Controls
// Fast low-res preview (F5) vs. smooth high-res final render (F6)
$fa = $preview ? 12 : 2;     // Minimum angle (degrees)
$fs = $preview ? 2 : 0.3;    // Minimum fragment size (mm)
$fn = 200;                     // Disabled so $fa and $fs take effect

// 3. BOSL2 Printer Tolerances
// Added clearance for 3D-printed joints, holes, and press-fits (in mm)
$slop = 0.2; 

// =================================================================
// CUSTOMIZER PARAMETERS
// =================================================================
/* [General Dimensions] */
renderType = "obj";
    globalSizerHeight = 10;
    
    boltHoleWidth = 15;
    
    
    pointHoleDown = 9;
    
    sizerRounding = 1;

    boltHoleGap = 2;
    
    boltHoleCount = 6;
    boltHoleFloor = 6;
    boltHoleHeight = 8;
    holderHeight = 9.8;
    boltHoleSideWallSize = 3;
    
//    boltHoleSize = [boltHoleWidth,boltHoleHeight,holderHeight];
    mode = "logo"; // logo box box-noText bottomRender
    
    bottomGapSize = 16;
    
    boltholeHeight = 8;
    
    pickHoleIn = -0.4;
    pickHoleDown = 3;
    
    pickRadius = 4.5;
    pickChamfer = 1;
    
    pickBottomRadius = 2.3;
    pickBottomHeight = 18;
    
    
    logoHeight = 40;
    pointBottomHeight = 13;
    pointFwdOffset = 0;
    textHeight = 1;
    
    textDepth = 0.7;
    
    textSize = 3; // 5;
    textShift = 3.5;//2.5;
    
    
    // logo
    
    logoScale = 0.63;
    logoCount = 7;
logoShiftFromPoint = -2*logoScale;

/* [Advanced Options] */
wall_thickness = 2.0;

    logoWedgeSideLength = 40;


// =================================================================
// DERIVED CONSTANTS & CALCULATIONS
// =================================================================
    globalSizerLenght = (boltHoleHeight+boltHoleGap)*boltHoleCount+boltHoleGap;
        globalSizerSize = [boltHoleWidth+boltHoleSideWallSize,globalSizerLenght,globalSizerHeight];

   

    centerBoxFwd = -2;
    centerBoxHeight = 1;
    centerBoxRaise = 3;
        centerBoxSize = [boltHoleWidth,globalSizerLenght+2+boltHoleSideWallSize,centerBoxHeight];
// =================================================================
// MODEL
// =================================================================
	include <BOSL2/std.scad>;


	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	
    


    
    
    
    module energy_smart(){
        pts = [[-16,-8],[-9,0],[-16,8],[-16,16],[-8,16],[0,8],[0,-8],[-8,-16],[-16,-16]];

        scale([logoScale,logoScale,1])
            rotate([0,0,-90])
            linear_extrude(height = holderHeight) {
                polygon(points = pts);
            }
    }


    module centerBox(feature=""){
    #fwd(centerBoxFwd)
            up(centerBoxRaise)
    cuboid(centerBoxSize, rounding=2, edges="Z");
    if(feature=="includeHeightBar"){

        right(20)
                cuboid([1,1,boltHoleFloor-textDepth], anchor=[-1,-1,-1]);
            
            }
    
    }
    module text_holes(height=textHeight, feature=""){
            ycopies(boltHoleHeight+boltHoleGap, n=boltHoleCount){

            fwd(textShift)
         //   down(2)
         up(boltHoleFloor-textDepth)
            text3d(str(($idx*10)+20), h=height,  anchor=BOT, size=textSize, font=":style=bold");
            
        }
        if(feature=="includeHeightBar"){
        up((boltHoleFloor-textDepth*2)/2)
        right(20)
                cuboid([1,1,boltHoleFloor-textDepth], anchor=[-1,-1,-1]);
            
            }
        
        
    }
    
    
            holeDepth = 12;
    module roundedHole(holeHeight=boltholeHeight, holeWidth=boltHoleWidth){
    cuboid([holeWidth,holeHeight,holeDepth/2], anchor=BOT);
    //, rounding=3, edges=BOTTOM, except=BACK
            up(holeDepth/2)
                       cuboid([holeWidth,holeHeight,holeDepth/2], anchor=BOT); //, rounding=-2, edges=TOP, except=BACK
    }
    
    module bolt_holes(boltHoleWidth=boltHoleWidth){
        ycopies(boltholeHeight+boltHoleGap, n=boltHoleCount){
        if($idx==0){
        down(0)
        fwd(pointHoleDown)
        intersection(){
       // scale([1,1,2])
             half_of([0, -1, 0], cp = -7){
            
           bottom_point(pointBottomHeight=pointBottomHeight*1.8);
            }
                roundedHole(holeHeight=boltholeHeight+3, holeWidth=boltHoleWidth);
           
            
        }
        }
           
       roundedHole(holeHeight=boltholeHeight, holeWidth=boltHoleWidth);
                       
                        
            }
    }
    
    module golf_pick_hole(){
        up(globalSizerHeight-pickHoleIn)
        fwd(pickHoleDown)
        ycyl(r=pickRadius, h=globalSizerLenght-pickBottomHeight/2+bottomGapSize/2, chamfer2=-pickChamfer, chamfer1=pickChamfer);
        //down(globalSizerLenght+bottomGapSize/2+5)
        if(pickBottomHeight > 0){
            up(globalSizerHeight-pickHoleIn)
            fwd((globalSizerLenght+bottomGapSize/2)/2)
             ycyl(r=pickBottomRadius, h=pickBottomHeight);
         }
    }
    
    bottomPointRadius = 21;
    bottomPointBaseAngleSpread = boltHoleWidth+10;
    bottomPointUpShift =11;
    bottomPointMove = [0,bottomPointUpShift-bottomPointRadius,0];
    
    module bottom_point(pointBottomHeight=pointBottomHeight){
    difference(){
        move(bottomPointMove)
        rotate([0,0,90])
         linear_extrude(h=holderHeight)
        arc(r=bottomPointRadius, angle=[-bottomPointBaseAngleSpread, bottomPointBaseAngleSpread], rounding=[1, 1, 1], wedge=true);
        
        fwd(pointBottomHeight)
        down(0.001)
            cuboid([50,20,20], anchor=BOT+CENTER);
        }
    }
    
    
    bottomText = "ENERGY";
    
    bottomText2 = "SMART";
    bottomTextSize = 7.5;
        bottomTextHeight = 2;
        
        bottomTextLeft = -4;
        bottomTextLeftBack = 2;
        bottomTextShiftUp = -1;

    module bottomRender(textSize=bottomTextSize, textHeight=bottomTextHeight){
    //fwd(-globalSizerLenght/2)
    
    back(bottomTextLeft)
    right(bottomTextShiftUp)
    up(textHeight-0.01)
    zrot(90){
        text3d(bottomText, textHeight, orient=DOWN, size=bottomTextSize, anchor=BOT+CENTER, font=":style=bold");
        
        fwd(10)
        right(bottomTextLeftBack)
      //  back(5)
        text3d(bottomText2, textHeight, orient=DOWN, size=bottomTextSize, anchor=BOT+CENTER, font=":style=bold", spacing=1.1);
        }
    
    }
    
    
    /*
module bottom_point_old(){
    pointWedgeSize = [22, pointBottomHeight, 3];
    eps = 0.01; // Small overlap to prevent z-fighting / preview artifacts
    wedgeRot = -12;
    difference(){
        // Base main cuboid
        cuboid([globalSizerSize[0], bottomGapSize, globalSizerSize[2]], 
               anchor=BOT, rounding=sizerRounding);
        
        // Left wedge pushed outward past the -X outer edge
        left(globalSizerSize[0]/2 + eps)
            rotate([0, 90+wedgeRot, 0])
            wedge(pointWedgeSize, anchor=BOT+CENTER);

        // Right wedge pushed outward past the +X outer edge
        right(globalSizerSize[0]/2 + eps)
            rotate([0, -90-wedgeRot, 0])
            wedge(pointWedgeSize, anchor=BOT+CENTER);
        
            
            fwd(pointBottomHeight)
            cuboid([50,20,20], anchor=BOT+CENTER);
            
        
       

    }
}*/

module energy_smart_copied(){

up(0)
            fwd(-logoShiftFromPoint)
            ycopies(16*logoScale, n=logoCount){
                energy_smart();
                }
                }


	module bolt_sizer(mode=mode){
    if(mode != "text" && mode != "bottomRender" && mode != "centerBox"){
		difference(){
        union(){
            if(mode == "box" || mode == "box-noText"){
                cuboid(globalSizerSize, anchor=BOT, rounding=sizerRounding);
            } else if(mode == "logo"){
            energy_smart_copied();
            
                /*back(30)
                cuboid([boltHoleWidth+boltHoleSideWallSize*(1+logoScale),3,holderHeight], anchor=CENTER+BOT);*/
            }else{
            echo("mode not set");
                }
                
           if(mode != "text"){
                fwd(globalSizerLenght/2+bottomGapSize/2-pointFwdOffset)
               bottom_point();
                }
            }
            
            if(mode != "logo"){
                up(boltHoleFloor)
                bolt_holes();
               // scale([0.4,1,1])
                //energy_smart_copied();
           } else {
           //     up(12)
           fwd(1)
                up(boltHoleFloor)
                bolt_holes(boltHoleWidth=8);
          }
            
            if(mode != "box-noText"){
                
                text_holes(height=textHeight);
                bottomRender();
            }
            
            if(mode != "logo"){
                golf_pick_hole();
            }
            
        }
          if(mode == "logo"){ 
                centerBox();
                }
	} else if(mode == "text") {
    down(textHeight)
            text_holes(height=textHeight, feature="includeHeightBar");
} else if(mode == "bottomRender"){
    down(bottomTextHeight)
    bottomRender(textHeight=bottomTextHeight*3);
    } else if(mode == "centerBox"){
                
                centerBox(feature="includeHeightBar");
                }
    }

   // sliced(renderType=renderType) {
   //back_half()

        bolt_sizer(mode=mode);  // }

//xcopies(80, n=5)
//energy_smart();

