include <BOSL2/std.scad>;

$fn = 100;
mode = "outer"; // outer inner all "footprint"
handleStyle = "handGrip"; // screwInHole screwMount


//  Screw handle Style
handGripSize = [10,20,30];

globalScale = [1,1,1];

// The clear opening is the trash chute. The inner tube holds the bag and
// nests inside the outer tube, which provides the handle and outside lip.
trashChuteDiameter = 195;
innerTubeWall = 2;
innerTubeHeightDifference = -2.5;

outerTubeWall = 1;
outerTubeHeight = 21;
innerTubeHeight = outerTubeHeight + innerTubeHeightDifference;

fitClearance = 0.35;
edgeRounding = -8.2;
edgeTopRounding = 1.35;

trashChuteRadius = trashChuteDiameter / 2;
innerTubeOuterRadius = trashChuteRadius + innerTubeWall;
outerTubeOuterRadius = innerTubeOuterRadius + outerTubeWall;

handleDepth = 60;
handleWidth = 110;
handleSize = [handleWidth, handleDepth, outerTubeHeight];
handleOverlap = trashChuteDiameter*0.1;

fingerGroveCount = 4;
fingerGrooveDiameter = 28;
fingerGroveYCopyGap = fingerGrooveDiameter+6;

fingerGrooveSpacing = fingerGrooveDiameter*0.7;


GroveOffset = 22;





connectorSpacing = 5;
connectorScaling=  [1,2,1.5];
connectorRadius = handleSize.z/2;

connectorCylCount = 20;
connectorCylChamfer1 = 26;
connectorSpacerInward = 0;
connectorSpacerMove = [0,-handleDepth/2-connectorSpacerInward,0];
connectorSpacerHeight = 30;



/// GLOBAL BOUNDS
globalBoundsHeight = 16;
globalBoundsMove = [20,20,-1.5];
globalBoundsBuffer = 15;
globalBounds = [248,248,globalBoundsHeight];
globalBoundsSpin = 45;
innerBoundingBoxUp = -2;
globalBoundsXYScale = 2;


footprintBounds = [248,248,2];


textRotate = [-140,0,0];
textOutOffset =5.9;
textMove = [0, trashChuteDiameter/2 + innerTubeWall + outerTubeWall + textOutOffset, 0];
textLabel = "BE THE CHANGE";
textSize= 8.8;
textHeight =1.5;


grapTabSize = [20,20,2];
grabTabInward = -10;
grapTabMove = [0, -trashChuteDiameter/2-grapTabSize.y-grabTabInward, globalBoundsHeight/2+innerBoundingBoxUp-innerTubeHeight+5.2];

edgeInnerBotRounding = -.5;

// This is the complete bag-retaining part.
module innerTube() {

difference(){
    tube(
        ir = trashChuteRadius,
        or = innerTubeOuterRadius,
        h = innerTubeHeight-8,
        rounding1 = edgeRounding,
        chamfer = edgeInnerBotRounding,
        irounding = edgeInnerBotRounding
    );
    
        cyl(r=trashChuteRadius, h=innerTubeHeight+0.01, chamfer1=2);
    }
    move(grapTabMove)
     cuboid(grapTabSize, rounding=4, edges="Z", except_edges=BACK, anchor=BOT);
   
}

// A slightly enlarged profile gives the two parts a printable slip fit.
module innerTubeCutout(innerTubeMode="withHole") {
    if(innerTubeMode == "fillCenter"){
  /*  tube(
        ir = 1,
        or = innerTubeOuterRadius + fitClearance,
        h = innerTubeHeight + 2,
        rounding1 = -edgeTopRounding + fitClearance,
        rounding2 = edgeRounding+ fitClearance,
     //   irounding = edgeRounding + fitClearance,
    );*/
    cyl(
        r=innerTubeOuterRadius + fitClearance,
        h = innerTubeHeight +0.6,
        chamfer1 = edgeRounding + fitClearance,
        rounding2 = edgeRounding+ fitClearance,
     );
    } else if(innerTubeMode == "withHole") {
   tube(
        ir = trashChuteRadius - fitClearance,
        or = innerTubeOuterRadius + fitClearance,
        h = innerTubeHeight + 2,
      //  chamfer2 = -edgeRounding + fitClearance,
           chamfer1 = edgeRounding + fitClearance,
        irounding = edgeRounding + fitClearance,
    );
    } else {
    echo(innerTubeMode);
        assert(false, "Invalid innerTubeMode found");
    }
}


// --- PARAMETERS ---
text_depth = 15;        // Thick chunky 3D depth to match the reference print
font_style = "Liberation Sans:style=Bold";
line_spacing = -1.5;    // Negative spacing forces lines to overlap and merge solid

holeOffsetX =8;
holeOffsetY =20.5;

holeHeight = handleSize.z;
holeMove = [handleSize.x/2-holeOffsetX, trashChuteDiameter/2+handleSize.y-holeOffsetY-0.001, 0];
holeRadius = 3;

module label(){
    move(textMove)
        rotate(textRotate)

     text3d(textLabel,size=textSize, h=textHeight, anchor=CENTER,  font = font_style);
    

}



module handle(handleStyle="handGrip") {
    if(handleStyle == "handGrip"){
        difference() {
            // Start inside the outer lip so this cuboid fuses cleanly to it.
            back(outerTubeOuterRadius + handleSize.y / 2 - handleOverlap)

                    cuboid(handleSize, rounding = 2, anchor=CENTER);
                    
        //            fwd(3)
        //            cuboid(handleSize, chamfer = -edgeRounding, edges=LEFT, anchor=BOT);
        
    //    fwd(10)
      /* move(connectorSpacerMove)
                    xcopies(connectorSpacing, connectorCylCount)
                    scale(connectorScaling)
                        ycyl(r=connectorRadius, h=connectorSpacerHeight, chamfer1=-connectorCylChamfer1, anchor=BOT);
       */            //  move(connectorSpacerBoundsMove)
        //            cuboid(connectorSpacerBounds, rounding = edgeRounding);
        
                

            // Four through-grooves give fingers a positive, comfortable grip.
            back(outerTubeOuterRadius + handleSize.y / 2 - handleOverlap + GroveOffset)
                ycopies(fingerGroveYCopyGap, 2)
                down(0.001)
                xcopies(spacing = fingerGrooveSpacing, n = fingerGroveCount)
                    
                    cyl(d = fingerGrooveDiameter, h = handleSize.z+0.03, rounding=-5, anchor=CENTER);
                   
                   if(holeRadius > 0){
               move(holeMove)
                zcyl(r=holeRadius, h=handleSize.z+0.001, anchor=CENTER, chamfer=-2);
                
                }

            // Keep the handle out of both the inner tube and trash chute.
            innerTubeCutout(innerTubeMode="fillCenter");
            
            
            
         //   label();
        }
    } else if(handleStyle == "screwMount"){
        cuboid(screwMountSize);
} else {
        assert(false, "handleStyle not valid");
    }
}

module outerGroove(square_side = 40, cyl_radius = 5, chamfer=-33){
groveOut =1.7;
// Define square size and cylinder dimensions
cyl_height = square_side;

// Calculate the half-distance from the center
h = square_side / 2;
groveZScale = 0.6;
groveScale = [1,1,groveZScale];

scale(groveScale)
// Generate 4 copies at the corners of the square
move_copies([[-h, -h], [h, -h], [h, h], [-h, h]])
    rot(90*$idx)
    left(-square_side/2)
    fwd(groveOut)
    xcyl(r = cyl_radius, h = cyl_height+5, anchor = BOTTOM, chamfer=chamfer, $fn=6);

/*
rect_tube(
h = outerCutoutTubeHeight*groveHeightPercent,
   // size = [2 * (trashChuteRadius - 10), 2 * (trashChuteRadius - 10), outerTubeHeight + 0.1],
    //            isize = [2 * (outerTubeOuterRadius + 3), 2 * (outerTubeOuterRadius + 3)],
            size = [2 * (outerTubeOuterRadius + outerTubeDepth+12), 2 * (outerTubeOuterRadius + outerTubeDepth + 15)],
            isize = [2 * (outerTubeOuterRadius + groveDepth), 2 * (outerTubeOuterRadius + groveDepth)],
       //     rounding = 16,
            rounding =3,
            //irounding2 = 11
    // No inner rounding needed unless specified for the inside cutout rim
);*/
}


module outerTube(innerHoleRadius=26) {

//intersection() {
diff(){
cuboid(
    size = [2 * (outerTubeOuterRadius + 10), 2 * (outerTubeOuterRadius + 10), outerTubeHeight],
    rounding = 5,       // Matches rounding = 5
    edges = [TOP, BOT]  // Applies only to top and bottom horizontal rims
)
tag("remove") prismoid(
    size1 = [2 * (trashChuteRadius - 1), 2 * (trashChuteRadius - 1)],
    size2 = [2 * (trashChuteRadius - 10), 2 * (trashChuteRadius - 10)],
    h = outerTubeHeight + 0.1,
    rounding = innerHoleRadius,
    anchor=CENTER
);


/*cuboid(
    size = [2 * (trashChuteRadius - 10), 2 * (trashChuteRadius - 10), outerTubeHeight + 0.1],
    rounding=innerHoleRadius, edges="Z"
    // No inner rounding needed unless specified for the inside cutout rim
);*/
/*
h = outerCutoutTubeHeight,
            size = [2 * (outerTubeOuterRadius + outerTubeDepth), 2 * (outerTubeOuterRadius + outerTubeDepth)],
            isize = [2 * (outerTubeOuterRadius + 3), 2 * (outerTubeOuterRadius + 3)],
            rounding = 16,
            irounding =16
*/

groveDepth = 5;
groveHeight = 5;
groveHeightPercent = 0.6;

tag("remove") down(groveHeight) outerGroove(square_side = outerTubeOuterRadius+outerTubeOuterRadius+20-groveDepth, cyl_radius = 8);

zrot(45)
tag("remove") down(groveHeight) outerGroove(square_side = outerTubeOuterRadius+outerTubeOuterRadius+50-groveDepth, cyl_radius = 8);
}
       
      /*  difference() {
        tube(
            ir = trashChuteRadius,
            or = outerTubeOuterRadius,
            h = outerTubeHeight,
            chamfer = edgeRounding,
            chamfer1 = -10
        );
        
    rect_tube(
            h = outerTubeHeight,
            size = [outerTubeOuterRadius * 2, outerTubeOuterRadius * 2],
            isize = [trashChuteRadius * 2, trashChuteRadius * 2],
            chamfer = edgeRounding
        );

        // The larger tube is simply the inner tube's outside lip.
//        down(2)
        innerTubeCutout();
    }*/
}

module with_bounding_box(size=globalBounds, move=globalBoundsMove, globalBoundsScale=[globalBoundsXYScale, globalBoundsXYScale, 1]) {
    // 1. Move the entire intersection group to the target destination
   
        rotate([180,0,0])
        intersection() {
            // 2. The bounding box constraint
            
            scale(globalBoundsScale)
            move(move)
            rotate([0,0,globalBoundsSpin])
           cuboid(size);
            
            // 3. The objects inside the module
            children();
    }
}


outerCutoutTubeHeight = 5;
outerTubeDepth = 15;

module outerCutoutTube(){
      /* rect_tube(
            ir = outerTubeOuterRadius+3,
            or = outerTubeOuterRadius+outerTubeDepth,
            h = outerCutoutTubeHeight,
            rounding = 1
        );*/
        rect_tube(
            h = outerCutoutTubeHeight,
            size = [2 * (outerTubeOuterRadius + outerTubeDepth), 2 * (outerTubeOuterRadius + outerTubeDepth)],
            isize = [2 * (outerTubeOuterRadius + 3), 2 * (outerTubeOuterRadius + 3)],
            rounding = 16,
            irounding =16
        );

}


module trash_bag_holder(mode = "all") {
 /*   if (mode == "inner") {
        with_bounding_box(move=[0,0,innerBoundingBoxUp])
        innerTube();
    } else if (mode == "outer") {*/
  with_bounding_box(size=globalBounds)
        union() {
        
            outerTube();
            difference(){
                handle(handleStyle="handGrip");
                outerCutoutTube();
            }
        }
   /* } else if (mode == "all") {
    with_bounding_box()
            innerTube();
        with_bounding_box()
        union() {
            outerTube();
            handle();
        }
    } else if (mode == "footprint-outer") {
    with_bounding_box(size=footprintBounds, globalBoundsScale=[1,1,1])
            innerTube();
        with_bounding_box()
        union() {
            outerTube();
            handle();
        }
    }*/
}

scale(globalScale)
trash_bag_holder(mode = mode);
