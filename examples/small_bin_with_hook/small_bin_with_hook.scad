include <BOSL2/std.scad>;
include <BOSL2/partitions.scad>;
include <BOSL2/joiners.scad>;
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

// Slicing configuration 
// "topHalf" or "bottomHalf" handles single-piece export masks natively via selective clipping.
partType = "hook"; //"halves"; // Options: "all", "container", "lid", "halves", "topHalf", "bottomHalf" "middleSection"
$slop = 0.15;        // Tolerances for 3D printing joint play

binRounding = 10;
binWallSize = 4;
binHookSideWidth = 80;
binNonHookSideWidgth = 80;
binHeight = 150;
lidHeight = 3;
binSize = [binHookSideWidth,binNonHookSideWidgth,binHeight];
lidSize = [binHookSideWidth,binNonHookSideWidgth,lidHeight];

hookStyle = "withHookGroove";
hookDepth = 15;
hookHeight = 12;
hookSize = [binHookSideWidth*0.8,hookHeight,hookDepth];
hookWallSize = 3;
hookRounding = 5;
hookSizeComputed = binHookSideWidth*0.8+hookWallSize;

lidHoleOffset = 12;
lidHoleOffsetVertical = 5;
lidHoleHingeLocation = -binNonHookSideWidgth/2+lidHoleOffset;
lidHoleMove = [lidHoleHingeLocation,  binHeight/2-lidHoleOffsetVertical, 0];
lidHoleDepth = 4;

module hook(){
    difference(){
        cuboid(hookSize+[hookWallSize, hookWallSize, hookWallSize], chamfer=hookRounding, edges="X", except_edges=TOP+BACK);
        down(hookWallSize/2) cuboid(hookSize+[20,0,1], chamfer=hookRounding, edges=[TOP,"Z",BOTTOM+RIGHT]); // Restored original fixed constant
    }
}
  
module small_bin_lid(){
    cuboid(lidSize, rounding=binRounding, edges="Z", anchor=CENTER);
    back(lidHoleHingeLocation) lidAxis();
    
    zrot(90)
    left(binNonHookSideWidgth/2-5)
    dovetail("male", width=5, height=5, slide=binHookSideWidth*0.8, chamfer=0.3);
}
  
module lidAxis(){
    xrot(90) yrot(90) cyl(r=lidHeight/2, h=binHookSideWidth+lidHoleDepth, rounding=1);
}

doveTailMove = [0,binNonHookSideWidgth/2,binHeight/2-20];
hookMountSize = [50,15,35];
slideJoinSlide = hookSizeComputed;
slideLargeWidth = 20;
dovetailHeight=4;


    module dovetailJoiner(type="male", slide=slide){
    rotate([0,90,90])
             dovetail(type, slide=slide, width=25, height=dovetailHeight, angle=20); //, back_width=slideLargeWidth
             }
             
module small_bin_with_hook(hookStyle = "includingHook"){
    difference(){
    union(){
            cuboid(binSize+[binWallSize,binWallSize,binWallSize], rounding=binRounding, except_edges=TOP, anchor=CENTER);
            up(binWallSize)
           
           
           // GROOVE MOUNT BLOCK
          if(hookStyle == "withHookGroove"){
          move(doveTailMove){
                    cuboid(hookMountSize, rounding=1);
            }
        }
  }  
     
                  up(binWallSize)
      cuboid(binSize, rounding=binRounding, except_edges=TOP, anchor=CENTER);
        move(lidHoleMove) lidAxis();
     
        // GROOVE
        if(hookStyle == "withHookGroove"){
       move(doveTailMove){

           // attach(TOP+LEFT){
            back(7.5)
            up(4)
                dovetailJoiner(type="female", slide=slideJoinSlide+30);
           // }
            }
        }
    }
    up(binHeight/2-hookHeight/2-hookWallSize/4)
    fwd(binNonHookSideWidgth/2+hookDepth/2)
    if(hookStyle == "includingHook"){ hook(); }
}// --- RENDER LAYER ---

if(partType == "all" || partType == "container"){
    small_bin_with_hook(hookStyle);
}

if(partType == "all" || partType == "lid"){
    small_bin_lid();
}

if(partType == "hook"){
            hook();
            back(hookDepth/2)
            dovetailJoiner(type="male", slide=slideJoinSlide);
            
}

// 1. Original logic handled cleanly without modification
if (partType == "halves" || partType == "topHalf" || partType == "bottomHalf") {
    // Renders the horizontal cut using a simple rotation technique
    xrot(-90)
    partition(
        size = binSize+[binWallSize,binWallSize,100],       // Box limits
        spread = ($preview && partType == "halves") ? 20 : 0, // Moves parts out during preview
        cutpath = "jigsaw",         // Interlocking profile shape
        cutsize = 12,                 // Size of dovetail locks
        gap = $slop                   // Snug tolerance offset
    )
    // Tumbles the container onto its side so the flat vertical cut divides it horizontal-ways
    xrot(90) {
        if (partType == "halves" || partType == "topHalf") {
            // Standard positive half space assignment
            right_half() small_bin_with_hook(hookStyle);
        }
        if (partType == "halves" || partType == "bottomHalf") {
            // Standard negative half space assignment 
            left_half() small_bin_with_hook(hookStyle);
        }
    }
}
    z_offset = binHeight / 6; // Where the cuts happen

    // Custom dimensions for your manual jigsaw teeth
    tab_width = 12;
    tab_height = 8;
    tab_thickness = binSize.y + binWallSize + 10; 
    module manual_jigsaw_cutter() {
        union() {
            // Main flat dividing wall (lower half block)
            translate([0, 0, -25]) 
                cuboid([binSize.x + 20, tab_thickness, 50], anchor=CENTER);
            
            // Loop across the X axis to generate interlocking teeth
            // Steps by double the width to leave precise spacing for the opposite side
            for (x = [-binSize.x/2 : tab_width*2 : binSize.x/2]) {
                translate([x + tab_width/2, 0, 0])
                    cuboid([tab_width, tab_thickness, tab_height*2], anchor=CENTER);
            }
        }
    }

// 2. Dedicated Middle Section Logic (Pure Library Approach)
// 2. Dedicated Middle Section Logic (Pure Library Approach)
if (partType == "middleSection") {
    z_offset = binHeight / 6; // Where the cuts happen
    
// Explicitly define the bounding dimension parameters
    mask_l = binSize.x + binWallSize + 1;
    mask_h = 100;

    difference() {
        // Start with the full container
        small_bin_with_hook(hookStyle);
        
        // 1. Library Call to remove the TOP cap section
        translate([0, 0, z_offset])
        xrot(-90)
        partition_cut_mask(l = mask_l,  h = mask_h, cutpath = "jigsaw", cutsize = 12, gap = $slop);

        // 2. Library Call to remove the BOTTOM floor section
        // Flipping it 180 degrees keeps the remaining middle piece intact
        translate([0, 0, -z_offset])
        rotate([0, 180, 0]) 
        xrot(-90)
        partition_cut_mask(l = mask_l,  h = mask_h, cutpath = "jigsaw", cutsize = 12, gap = $slop);
    }
}