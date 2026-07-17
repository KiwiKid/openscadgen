include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

outerPipeRadius = 7;
widerPipeRadius = 10;

femalePipeRadius = 5.5;
middlePipeRadius = femalePipeRadius + 3;

pipeHeight  = 40;
pipeHeadHeight = 10;
middleSectionHeight = 0.1;
holeDepth = pipeHeight - 8;

clipStyle = "clip";


module middleSection(){
difference(){
    // Changed to a simple cylinder bridging the two ends
    cyl(r=femalePipeRadius, h=middleSectionHeight, anchor=CENTER);
        //#text3d(str(middleSectionHeight));
         rotate([0,90,0])
         move([0,-femalePipeRadius/2-1,femalePipeRadius])
            text3d(str(middleSectionHeight), size=8, h=3, font="Liberation Sans");
        }
    
    
}

module maleEnd(){
    // Anchored to the TOP so its top aligns with Z=0, extending downwards
    cyl(r=femalePipeRadius, h=pipeHeadHeight, rounding1=3, anchor=TOP);
}

module femaleEnd(){
    // Anchored to the BOTTOM so its bottom aligns with Z=0, extending upwards
    difference(){
        cyl(r1=middlePipeRadius, r2=widerPipeRadius, h=pipeHeight, rounding=2, anchor=BOTTOM);
        
        // Internal hole cut out from the top down
        up(pipeHeight - holeDepth)
            cyl(r=outerPipeRadius, h=holeDepth + 0.01, rounding2=-3, anchor=BOTTOM);
    }
}


clipWidth = 2;
clipHeight = 40;
clipSize = [widerPipeRadius,clipWidth,clipHeight];
clipOffset = 1;

wedgeOut = 8;
clipConnectorWidth = 3;

clipConnectorHeight = 15;
clipDown = clipHeight-middlePipeRadius-10;
clipConnectorUp = clipHeight-clipConnectorHeight;

module clip(clipStyle){
    up(clipConnectorUp)
    fwd(widerPipeRadius-clipConnectorWidth/2)
    cuboid([widerPipeRadius,clipConnectorWidth,clipConnectorHeight], rounding=1,  edges="Y", anchor=BOTTOM);

    fwd(widerPipeRadius-clipWidth/2+clipOffset)
    cuboid(clipSize, rounding=1, edges="Y", anchor=BOTTOM)
    position(BOTTOM){
  //  back(wedgeOut/2-clipWidth)
  //  up(wedgeOut/2)
    xrot(270)
    back(0.5)
    scale([1,1,1.5])
        top_half()
        wedge(wedgeOut,10);
        }
}

module pipe_extender(){
    // 1. Male End at the bottom (extends from 0 downwards)
    
    down(middleSectionHeight/2)
    maleEnd();
    //middleSectionHeight
    // 2. Middle Section sits exactly on top of the male end
   
        middleSection();
    
    // 3. Female End sits on top of the middle section
     up(middleSectionHeight/2)
        femaleEnd();
        
        down(clipDown)
        clip(clipStyle=clipStyle);
}

pipe_extender();