
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

holderLength = 70;

holderWidth = 30;


holderHeight = 17.65;
holderRoundingReduction = 0.72;

spitHeightRatio = 1.3;
spoonSplitThickness = 4;


spoonSplitThickness2 = 1.5;
spoonSplitRotate2 = 11;

spoonSplitRotate = 11.5;
spoonSplitWidthRatio = 0.6;

isSizer = false;

spoonSplitSize = [spoonSplitThickness,holderWidth*spoonSplitWidthRatio,holderHeight*spitHeightRatio];

spoonSplitSize2 = [spoonSplitThickness2,holderWidth*spoonSplitWidthRatio,holderHeight*spitHeightRatio];

spoonSplitTranslate = [-holderLength*0.3,0,-1.1];


spoonSplitTranslate2 = [-holderLength*0.2,0,-1.1];


module spoonSlot(slotSize=slotSize, rot=slotRotate){
    yrot(rot){
        cuboid(slotSize, rounding=-8, edges=TOP);
        
//         spoon tester:
 //      scale([1,1,10])
 //      #cuboid(spoonSplitSize*2.2);
        }

}

module spoon_holder(){
   difference(){
   
	cuboid([holderLength,holderWidth, holderHeight], rounding=holderHeight*holderRoundingReduction, edges=TOP);
    
    
    translate(spoonSplitTranslate)
    spoonSlot(slotSize=spoonSplitSize, rot=spoonSplitRotate);
    
    //    translate(spoonSplitTranslate2)
  //  spoonSlot(slotSize=spoonSplitSize2, rot=spoonSplitRotate);
       

    }
    
}

if(isSizer == true){

intersection(){
spoon_holder();
translate([-10,0,0])
cuboid([15,holderWidth*0.7,100])
attach(RIGHT){

down(0.35)
text3d(str(spoonSplitThickness), center=true);
}

}

}else {
spoon_holder();
}
