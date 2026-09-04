
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

/* [Render Controls] */
// Select the generated object: box, box-noText, text, bottomRenderBox, bottomRender, centerBox, logo, glueJig, glueJoinerSquare, or all.
mode = "text"; //bottomRenderBoxHeight
textMode = "text"; // Set to "noText" to omit the top labels.
renderType = "obj";

// Thickness of the bottom slice that receives the image-relief replacement.
bottomRenderBoxHeight = 2;

/* [Glue Jig] */
// Thin alignment frame used while gluing the bottom insert into the sizer.
glueJigCount = 5;
glueJigSeparation = 10;
glueJigHeight = 10.99;
glueJigWidth = 14;
glueJigFloorHeight = 3;
glueJigBorder = 6;
glueJigLengthExtra = 8;

glueSquareCutoutMove = 13;
glueSquareCutoutSize = [2,40,1.5];
glueSquareCutoutChamfer = 0.3;

/* [Sizer Body] */
// Overall body dimensions and the pointed end attached to it.
globalSizerHeight = 10;
sizerRounding = 1;
bottomGapSize = 16;

pointHoleDown = 9;
pointBottomHeight = 12;
pointFwdOffset = 0;

/* [Bolt Holes] */
// Repeated bolt-hole cut-outs along the length of the sizer.
boltHoleWidth = 15;
boltHoleGap = 2;
boltHoleCount = 6;
boltHoleFloor = 3;
boltHoleHeight = 8;
boltholeHeight = 8;
boltHoleSideWallSize = 3;
holeDepth = 12;

/* [Golf Pick Hole] */
// Long relief cut and its larger end clearance.
pickHoleDown = 3;
pickRadius = 6.5;
pickHoleIn = -boltHoleFloor + 2;
pickChamfer = 5;
pickBottomRadius = 3.5;
pickBottomHeight = 18;

/* [Top Labels] */
// Raised or cut numeric labels above the bolt holes.
textHeight = 0.3;
// Thickness of the standalone text colour-modifier mesh.
textColorThickness = 0.3;
// Keeps the modifier just inside the label surface and avoids coincident faces.
textModifierSurfaceOffset = 0.01;
// Lowers the standalone text modifier while preserving its Z=0 support bar.
textModifierDown = 0.15;
textDepth = 0.7;
textSize = 5;
textShift = 3.5;

/* [Bottom Image Relief] */
// Height-map relief generated from image.png on the bottom of the sizer.
bottomRenderUp = 2;
bottomRenderTextHeight = 2;
bottomRenderModifierZeroHeight = 1.15;
bottomRenderModifierHeight = -bottomRenderModifierZeroHeight-0;
// Cut the image-relief cavity into the sliced bottom box.
bottomRenderBoxWithText = true;

logoScaleAdjust = 0.53;
logoShiftRight = 0;

/* [Logo Geometry] */
// Settings used by the optional Energy Smart logo modules.
logoScale = 0.28;
logoCount = 7;
logoHeight = 1;
logoShiftFromPoint = -4 * logoScale;
logoWedgeSideLength = 40;

/* [Advanced Options] */
// Reserved for wall-thickness based features.
wall_thickness = 2.0;


// =================================================================
// DERIVED CONSTANTS & CALCULATIONS
// =================================================================
globalSizerLenght = (boltHoleHeight + boltHoleGap) * boltHoleCount + boltHoleGap;
globalSizerSize = [boltHoleWidth + boltHoleSideWallSize, globalSizerLenght, globalSizerHeight];

// Pointed end geometry, derived from the body and bolt-hole dimensions.
bottomPointRadius = 21;
bottomPointBaseAngleSpread = boltHoleWidth + 10;
bottomPointUpShift = 11;
bottomPointMove = [0, bottomPointUpShift - bottomPointRadius, 0];




roud = 1;
// =================================================================
// MODEL
//


	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/







    module energy_smart_img(height = logoHeight, logoScale=logoScale){

      // pts = [[-16,-8],[-9,0],[-16,8],[-16,16],[-8,16],[0,8],[0,-8],[-8,-16],[-16,-16]];
        linear_extrude(h=logoHeight)
    rotate([0,180,0])
        scale([logoScale,logoScale,1])
            rotate([0,0,-90])


            projection(cut = true) {
        // Translate the surface down on the Z-axis to choose where the 2D cut happens
        translate([-logoShiftFromPoint, 0, -10])
                // 1. Define image dimensions (X, Y, and maximum extruded Z height)

        // 2. Wrap it in a BOSL2 attachable block
        attachable(anchor=CENTER, spin=0) {

            // Standard OpenSCAD surface call centered to match BOSL2 bounds
            scale([1, 1, 1]) // Adjust scaling if necessary
             surface("image.png",  center = true);
          //  surface(file="heightmap.png", center=true, invert=true);

            // Explicitly declare attachments anchor position (internal mechanic)
            children();
        }


    }
}



    module centerBox(feature=""){
    fwd(centerBoxFwd)
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
           down(textHeight-0.7)
         up(boltHoleFloor-textDepth)
            text3d(str(($idx*10)+20), h=height,  anchor=BOT, size=textSize, font=":style=bold");

        }
        if(feature=="includeHeightBar"){
        up((boltHoleFloor-textDepth*2)/2)
        right(20)
                cuboid([1,1,boltHoleFloor-textDepth], anchor=[-1,-1,-1]);

            }


    }

    // The standalone colour modifier must overlap the matching text recess.
    // If it is fully below or above that recess, Bambu Studio cannot apply it.
    module assertTextModifierIntersectsTextRecess() {
        textRecessBottom = boltHoleFloor - textDepth - (textHeight - 0.7);
        textRecessTop = textRecessBottom + textHeight;
        textModifierBottom = textRecessBottom - textModifierSurfaceOffset - textModifierDown;
        textModifierTop = textModifierBottom + textColorThickness;

        assert(
            textModifierTop > textRecessBottom,
            str(
                "Text modifier is too low: its top (", textModifierTop,
                ") does not reach the text recess bottom (", textRecessBottom, ")."
            )
        );
        assert(
            textModifierBottom < textRecessTop,
            str(
                "Text modifier is too high: its bottom (", textModifierBottom,
                ") is above the text recess top (", textRecessTop, ")."
            )
        );
    }


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
        ycyl(r=pickRadius, h=globalSizerLenght-pickBottomHeight/2+bottomGapSize/2, chamfer1=pickChamfer);
        //down(globalSizerLenght+bottomGapSize/2+5)
        if(pickBottomHeight > 0){
//            up(globalSizerHeight-pickHoleIn)
            up(globalSizerHeight)
            fwd((globalSizerLenght+bottomGapSize/2)/2)
             ycyl(r=pickBottomRadius, h=pickBottomHeight);
         }
    }

    module bottom_point(pointBottomHeight=pointBottomHeight){
    difference(){
        move(bottomPointMove)
        rotate([0,0,90])
         linear_extrude(h=globalSizerHeight)
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

        bottomTextLeft = -10;
        bottomTextLeftBack = 2;
        bottomTextShiftUp = -1;
/*
    module bottomRender(textHeight=10, textDepth = 4, logoScaleAdjust=logoScaleAdjust){

      //  scale([1,1,4])
      //  up(textDepth/4.5)
   //   down(-1)
   up(bottomRenderModifierHeight)
         zcopies(0.1,textDepth){

           fwd(logoShiftRight)
            difference(){
           down(0.4)
            cuboid([16,globalSizerLenght,0.4], anchor=BOT);
            //fwd(-globalSizerLenght/2)

            up(0.1)
            rotate([0,180,90])
            resize([0, 0, logoScaleAdjust], auto=[true, true, false]) {
                linear_extrude(height = textHeight, convexity = 10) {
                    projection(cut = true) {
                        // Adjust the Z-translation to slice right through the middle of your text
                        translate([0, 0, 80]) {
                            surface(file = "image.png", center = true, invert = true);
                        }
                    }
                }
            }

            }
        }
    }*/


    module logoProjection(height = 10, scaleZ = logoScaleAdjust) {
    rotate([0, 180, 90])
        resize([0, 0, scaleZ], auto = [true, true, false]) {
            linear_extrude(height = height, convexity = 10) {
                projection(cut = true) {
                    translate([0, 0, 80]) {
                        surface(file = "image.png", center = true, invert = true);
                    }
                }
            }
        }
}

// Main rendering module
module bottomRender(textHeight = 10, textDepth = 4, logoScaleAdjust = logoScaleAdjust, cutoutOffset = 0.2) {
    up(bottomRenderModifierHeight)
        zcopies(0.1, textDepth) {
            fwd(logoShiftRight)
                difference() {
                    // Main body
                    down(0.4)
                        cuboid([16, globalSizerLenght, 0.4], anchor = BOT);

                    // Primary logo cutout
                    up(0.1)
                        logoProjection(height = textHeight, scaleZ = logoScaleAdjust);

                    // Secondary offset cutout
                    down(0.1)
                        logoProjection(height = textHeight, scaleZ = logoScaleAdjust);
                }
        }
}

// Position the relief to overlap the bottom slice before it is flipped.
module bottomRenderAtFloor() {
    up(0.4)
        bottomRender(textHeight = bottomRenderTextHeight);
}

// The lower slice removed from the main box and exported as the logo base.
module bottomRenderBoxBase() {
    intersection() {
        bolt_sizer(mode = "box", textMode = "noText");
        up(bottomRenderBoxHeight / 2)
            cuboid([100, 100, bottomRenderBoxHeight]);
    }
}

// The sliced base and matching relief cavity used by the logo export.
module bottomRenderBoxCutout(withText = true) {
    difference() {
        bottomRenderBoxBase();
        if (withText)
            bottomRenderAtFloor();

            glueSquareCutout(cutoutUp=0);
    }
}

module glueSquareCutout(cutoutUp=bottomRenderBoxHeight){
up(bottomRenderBoxHeight)
left(0)
//xcopies(glueSquareCutoutMove, 2)
 //   cuboid(glueSquareCutoutSize, chamfer=glueSquareCutoutChamfer);

    if(cutoutUp > 0){
    left(20)
    cuboid([1,1,10], anchor=BOT);
    }
}

// Main sizer body with space for the separately printed bottom slice.
module boxWithBottomRenderCutout() {
    difference() {
        bolt_sizer(mode = "box");
        bottomRenderBoxBase();

      //  glueSquareCutout();
    }
}

// The completed sizer: main body plus the bottom-render insert in its recess.
module boxWithBottomRenderInsert() {
    union() {
        boxWithBottomRenderCutout();
        bottomRenderBoxCutout(withText = true);
    }
}

glueJigSize = [
        //globalSizerSize[0] + 2 * glueJigBorder,
         20,
        globalSizerSize[1] + glueJigLengthExtra + 2 * glueJigBorder,
        glueJigHeight,
    ];
module glueJigCutout(){
    up(glueJigHeight - 1)
        fwd(5)
            xcopies(glueJigSize[1], n = 2) {
                cuboid([1, glueJigSize[1] - 5, 10]);
            }
}

// Separate, slightly undersized strips that sit in the matching glue channels
// between the box and bottom insert. Keep this geometry in sync with
// glueSquareCutout().
module glueJoinerSquare(size = glueSquareCutoutSize * 0.96) {
   /* xcopies(glueSquareCutoutMove, n = 2)
        cuboid(size, chamfer = glueSquareCutoutChamfer, anchor = BOT);*/
}

// One thin alignment frame with the box outline and glue-key channels removed.
module glueJigFrame() {
    difference() {
        fwd(glueJigLengthExtra / 2)
            cuboid(glueJigSize, anchor = BOT, chamfer=roud);

        // Cut the full assembled sizer, including its bottom-render insert.
        up(glueJigFloorHeight)
        down(bottomRenderBoxHeight)
            hull()
                boxWithBottomRenderInsert();
         
         
         up(glueJigFloorHeight)
         cuboid([25,55,10], anchor = BOT,);
         
         
      //   up(glueJigFloorHeight-5)
         

    }
}
ridgesHeight = 3.5;
// Print one alignment frame per sizer, arranged in a row for batch gluing.
module glueJig() {
difference(){
union(){
    xcopies(globalSizerSize[0] + glueJigSeparation, n = glueJigCount)
        glueJigFrame();
       
       
       fwd(5)
      ycopies(globalSizerSize[1]+17,2)
        cuboid([(globalSizerSize[0]+glueJigSeparation)* glueJigCount-20, 6, 3], anchor=BOT, chamfer=roud);
        
        }
        fwd(5)
        up(ridgesHeight)
        xcopies(3,45)
            cuboid([1,globalSizerSize[1]+15,15], anchor=BOT);
            
            down(2)
            cuboid([120,55,20], anchor = BOT,);
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
}

module energy_smart_copied(){

up(0)
            fwd(-logoShiftFromPoint)
            ycopies(16*logoScale, n=logoCount){
                energy_smart();
                }
                }
*/
module bottomLogo(){
back(globalSizerSize[1]/2-10){
                energy_smart(height=textHeight*2, logoScale=0.55);
                fwd(8.8)
                energy_smart(height=textHeight*2, logoScale=0.55);
                }
}

	module bolt_sizer(mode=mode, textMode="text"){
    if(mode != "text" && mode != "glueJoinerSquare" && mode != "bottomRender" && mode != "centerBox"){
		difference(){
        union(){
            if(mode == "box" || mode == "box-noText" || mode=="logo"){
                cuboid(globalSizerSize, anchor=BOT, rounding=sizerRounding);



            } else if(mode == "logo"){
           // up(logoHeight)
             //   energy_smart_img();

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

        //    bottomLogo();

       //     if(mode != "logo"){
                up(boltHoleFloor)
                bolt_holes();
               // scale([0.4,1,1])
                //energy_smart_copied();
    //       } else {
           //     up(12)
     //      fwd(1)
      //          up(boltHoleFloor)
       //         bolt_holes(boltHoleWidth=8);
       //   }

            if(textMode != "noText"){

              text_holes(height=textHeight);
              bottomRender(textHeight=bottomRenderTextHeight, textDepth=bottomRenderUp);

            }

         //   if(mode != "logo"){
                golf_pick_hole();
        //    }

        }
       /*   if(mode == "logo"){
                #centerBox();
               }*/
	} else if(mode == "text") {
    // Keep the letters at the box's label recess, just below its surface.
    // The off-body support bar starts at Z = 0, preventing Bambu Studio from
    // dropping this modifier to the build plate during import.
    assertTextModifierIntersectsTextRecess();
    union() {
        down(textModifierSurfaceOffset + textModifierDown)
            text_holes(height=textColorThickness);
        right(globalSizerSize[0] / 2 + 2)
            cuboid([1, 1, boltHoleFloor - textModifierSurfaceOffset - textModifierDown], anchor = BOT);
    }
} else if(mode == "bottomRender"){

   // Match the flipped logo cutout so the relief is ready to print face-up.
   down(1)
    rotate([180,0,0])
      bottomRenderAtFloor();

    } else if(mode == "centerBox"){

                centerBox(feature="includeHeightBar");
                }
    }

/*
sizerHeight = 10;
        fwd(globalSizerLenght/2+pointBottomHeight-7)
        ycopies(10,7, sp=[0,0,0]){
            cuboid([1,sizerHeight,20]);
        }*/
       echo(mode);
   if(mode == "all"){
        boxWithBottomRenderCutout();

       #bolt_sizer(mode="text");
       bolt_sizer(mode="bottomRender");

  } else if(mode == "box"){
    boxWithBottomRenderCutout();

  } else if(mode == "glueJig"){
    glueJig();

  } else if(mode == "glueJoinerSquare"){
    glueJoinerSquare();

  } else if(mode == "logo"){
    // Export the face-up bottom slice with the cavity for bottomRender.
    rotate([180,0,0])
      bottomRenderBoxCutout(withText = true);

  } else if(mode == "bottomRenderBox"){
    // Preview/export the same face-up slice, optionally without the cavity.
    rotate([180,0,0])
      bottomRenderBoxCutout(withText = bottomRenderBoxWithText);

   } else{
    bolt_sizer(mode=mode);
  }
