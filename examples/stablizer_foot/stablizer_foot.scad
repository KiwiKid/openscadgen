
include <BOSL2/std.scad>;
include <BOSL2/isosurface.scad>
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

footHeight = 50;
footRadius = 60;
topRadius = 20;

lConnectorWidth = 3.5;
lConnectorLength = 35; 
lConnectorHeight = footHeight+100;
LConnectorDepth = 10;
LConnectorWidth = 5;
LConnectorEndSize = [lConnectorHeight,LConnectorDepth,LConnectorWidth];

lCenterOffset = 1;

outerCubeSize = [1000,1000,150];

footRoundingCutoff = 50;

/*
renderType:
use to print a test slice and confirm sizing before printing:
 - "horzSlice" - horizontal slices (default)
 - "vertSlice" - vertical slices
 - "all" - the whole object
*/
renderType = "obj";

assert(LConnectorDepth < footHeight, "LConnectorDepth too large, must be less than footHeight");


connectorType = "L"; // LnoBlocks

module lSection(LConnectorEndSize=LConnectorEndSize){
    cuboid([lConnectorLength, lConnectorHeight, lConnectorWidth], anchor=BOT)
        attach(LEFT){
            fwd(LConnectorDepth/2-6)
            cuboid(LConnectorEndSize, anchor=BOT);
        }
       
}


module hole(connectorType=connectorType){
if(connectorType == "L"){
    right(lConnectorLength/2)
    up(LConnectorDepth+lConnectorHeight/2+footRoundingCutoff){
        //cuboid([lConnectorLength, lConnectorWidth, lConnectorHeight])
        rotate([90,180,0]){
        right(lCenterOffset)
        lSection(LConnectorEndSize=LConnectorEndSize);
        
        
        
        right(lConnectorLength/2)
        down(lConnectorLength/2-lCenterOffset)
        rotate([180,-90,0])
         lSection(LConnectorEndSize=LConnectorEndSize);
         }
            
        
        }
    }
    
    if(connectorType == "LnoBlocks"){
        right(lConnectorLength/2)
        up(LConnectorDepth+lConnectorLength/2+footRoundingCutoff){
            //cuboid([lConnectorLength, lConnectorWidth, lConnectorHeight])
            rotate([90,180,0]){
            right(lCenterOffset)
            lSection(LConnectorEndSize=[0,0,0]);
            
            
            
            right(lConnectorLength/2)
            down(lConnectorLength/2-lCenterOffset)
            rotate([180,-90,0])
             lSection(LConnectorEndSize=[0,0,0]);
             }
                
            
            }
    }
}

module stablizer_foot(){
intersection(){
cuboid(outerCubeSize, anchor=BOT);
down(footRoundingCutoff)
difference(){
    snowMan(scaleFactor=2);
	//cyl(r1=footRadius, r2=topRadius, h=footHeight);
       hole(connectorType=connectorType);
    
    }
}
}


module snowMan(scaleFactor=1, large_radius=20 ,small_radius=8, separation=12){

// Dimensions
//large_radius = 20;
//small_radius = 8;
//separation   = 12; // Gap distance between the two ball surfaces

// Ground alignment calculation
large_z = large_radius; 
small_z = (large_radius * 2) + separation + small_radius;

// Flat 1D list pairing transform matrices with function objects
spec = [
    // Matrix transform                // Metaball Shape Function
    xrot(0) * up(large_z),             mb_sphere(r=large_radius), 
    xrot(0) * up(small_z),             mb_sphere(r=small_radius)
];

// Generate the interacting metaballs mesh
scale(scaleFactor)
metaballs(
    spec = spec, 
    bounding_box = [ [-25, -25, -2], [25, 25, small_z + small_radius + 5] ], 
    voxel_size = 1.0  
);
}







module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 1.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, -16],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}



  sliced(renderType=renderType) {
        stablizer_foot();
    }